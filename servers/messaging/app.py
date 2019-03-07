from flask import Flask, request, Response
import json
from models import Member, Message, Channel, db
from datetime import datetime, timezone
import config
import pika
import sys

flask_app = Flask(__name__)
flask_app.config["SQLALCHEMY_DATABASE_URI"] = config.DATABASE_CONNECTION_URI
flask_app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = False
flask_app.app_context().push()
db.init_app(flask_app)
db.create_all()
db.session.add(Channel(name='general'))
db.session.commit()

# RabbitMQ declaration
creds = pika.PlainCredentials(config.rUSER, config.rPW)
conn = pika.BlockingConnection(pika.ConnectionParameters(host=config.mqHOST, port=config.mqPORT, credentials=creds, heartbeat=0))
mq_chan = conn.channel()
mq_chan.queue_declare(queue=config.rmQueue, durable=True)


@flask_app.route("/v1/channels", methods=['GET', 'POST'])
def ChannelHandler():
    if 'X-User' not in request.headers:
        return Response('Unauthorized Access', 401, mimetype='text/html')
    data = json.loads(request.headers['X-User'])
    user = find_user(data["id"])
    if request.method == 'GET':
        try:
            channs = (db.session.query(Channel).filter((Channel.members.any(id=data['id'])) | (Channel.private==False)).all())
            return Response(json.dumps([c.as_dict() for c in channs]), 200, mimetype='application/json')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')
            

    if request.method == 'POST':
        chann = Channel(**request.json)
        if chann.name == None:
            return Response('Bad Request', 400, mimetype='text/html')

        if chann.members != None:
            for m in chann.members:
                find_user(m.id)

        try:
            # add channel
            chann.creator = user
            db.session.add(chann)
            db.session.commit()
            # send event to msg queue
            event = {}
            event["type"] = "channel-new"
            event["channel"] = chann.as_dict()
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))
            return Response(json.dumps(chann.as_dict()), 201, mimetype='application/json')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')


@flask_app.route("/v1/channels/<channel_id>", methods=['GET', 'POST', 'PATCH', 'DELETE'])
def SpecificChannelHandler(channel_id):
    if 'X-User' not in request.headers:
        return Response('Unauthorized Access', 401, mimetype='text/html')
    
    chann = db.session.query(Channel).get(channel_id)
    user = json.loads(request.headers['X-User'])
    user = find_user(user["id"])
    if request.method == 'GET':
        # check if private channel and user apart of channel
        if chann.private==True and user not in chann.members:
            return Response('Access Forbidden', 403, mimetype='text/html')
        
        # check if before query parameter
        before = request.args.get('before')
        if before:
            # get 100 messages before this message id
            msgs = db.session.query(Message).filter((Message.chann_id==chann.id) & (Message.id < before)).limit(100).all()
            return Response(json.dumps([m.as_dict() for m in msgs]), 200, mimetype='application/json')
        else:
            # get 100 most recent messages
            msgs = db.session.query(Message).filter(Message.chann_id==chann.id).order_by(Message.id.desc()).limit(100)
            return Response(json.dumps([m.as_dict() for m in msgs]), 200, mimetype='application/json')

    if request.method == 'POST':
        if chann.private==True and user not in chann.members:
            return Response('Access Forbidden', 403, 'text/html')
        
        # create new message with json body and add to db
        # respond with 201, application/json and copy of new message as json
        msg = Message(body=request.json['body'], chann_id=chann.id)
        try:
            msg.creator = user
            db.session.add(msg)
            db.session.commit()

             # send event to msg queue
            event = {}
            event["type"] = "message-new"
            event["message"] = msg.as_dict()
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))

            return Response(json.dumps(msg.as_dict()), 201, mimetype='application/json')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')


    if request.method == 'PATCH':
        # check user is creator of channel
        if user.id != chann.creator_id:
            return Response('Access Forbidden', 403, mimetype='text/html')
        
        name = request.json['name']
        descr = request.json['description']
        try:
            if name != None:
                chann.name = name
            
            if descr != None:
                chann.description = descr
            chann.edited_at = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S")
            db.session.commit()

             # send event to msg queue
            event = {}
            event["type"] = "channel-update"
            event["channel"] = chann.as_dict()
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))

            return Response(json.dumps(chann.as_dict()), 200, mimetype='application/json')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')
    

    if request.method == 'DELETE':
        # check user is creator
        if user.id != chann.creator_id:
            return Response('Access Forbidden', 403, mimetype='text/html')
        
        # delete channel from db and all messages related to channel
        # respond with text/html indicating delete success
        try:
            Message.query.filter_by(chann_id=chann.id).delete()
            Channel.query.filter_by(id=chann.id).delete()
            db.session.commit()

             # send event to msg queue
            event = {}
            event["type"] = "channel-delete"
            event["channelID"] = chann.id
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))

            return Response('Channel Delete Successful', 200, mimetype='text/html')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Internal Server Error', 500, mimetype='text/html')
    

@flask_app.route("/v1/channels/<channel_id>/members", methods=['POST', 'DELETE'])
def ChannelMembersHandler(channel_id):
    if 'X-User' not in request.headers:
        return Response('Unauthorized Access', 401, mimetype='text/html')
    
    user = json.loads(request.headers['X-User'])
    user = find_user(user["id"])
    chann = db.session.query(Channel).get(channel_id)

    if chann == None:
        return Response('Bad Request', 400, mimetype='text/html')

    if user.id != chann.creator_id:
        return Response('Access Forbidden', 403, mimetype='text/html')
    
    if request.method == 'POST':
        #nu = request.json
        try:
            m = Member(**request.json)
            # add user to member table to save data for future usage
            if Member.query.filter_by(id=m.id).first() == None:
                db.session.add(m)
                db.session.commit()

            # add to association table the new relationship b/w channel and member
            q = 'INSERT INTO chann_mems (member_id, channel_id) VALUES (%s, %s)'
            db.engine.execute(q, m.id, chann.id)
            db.session.commit()
            return Response('User Added', 201, mimetype='text/html')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')

    
    if request.method == 'DELETE':
        rm = request.json
        try:
            q = 'DELETE FROM chann_mems WHERE member_id=%s'
            db.engine.execute(q, rm['id'])
            db.session.commit()
            return Response('User Deleted', 200, mimetype='text/html')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')


@flask_app.route("/v1/messages/<message_id>", methods=['PATCH', 'DELETE'])
def MessagesHandler(message_id):
    if 'X-User' not in request.headers:
        return Response('Unauthorized Access', 401, mimetype='text/html')
    
    user = json.loads(request.headers['X-User'])
    user = find_user(user["id"])
    msg = Message.query.filter_by(id=message_id).first()

    if user.id != msg.creator_id:
        return Response('Access Forbidden', 403, mimetype='text/html')

    if request.method == 'PATCH':
        # update message body property
        # return updated msg, encoded as json, application/json
        try:
            msg.body = request.json['body']
            msg.edited_at = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S")
            db.session.commit()

             # send event to msg queue
            event = {}
            event["type"] = "message-update"
            event["message"] = msg.as_dict()
            chann = db.session.query(Channel).get(msg.chann_id)
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))

            return Response(json.dumps(msg.as_dict()), 200, mimetype='application/json')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')

    if request.method == 'DELETE':
        # delete message, text/html that successful delete
        try:
            Message.query.filter(Message.id==msg.id).delete()
            db.session.commit()

            # send event to msg queue
            event = {}
            event["type"] = "message-delete"
            event["messageID"] = msg.id
            chann = db.session.query(Channel).get(msg.chann_id)
            if chann.private == True:
                event["userIDs"] = [m.id for m in chann.members]
            mq_chan.basic_publish(exchange='', routing_key=config.rmQueue, body=json.dumps(event))

            return Response('Delete Successful', 200, mimetype='text/html')
        except Exception as e:
            print(e, file=sys.stderr)
            db.session.rollback()
            return Response('Bad Request', 400, mimetype='text/html')


# find_user takes the id and finds the user in the users table
# and adds it to the member table 
def find_user(id):
    if Member.query.filter_by(id=id).first() == None:
        try:
            q = "insert into member (id, userName, firstName, lastName, photoURL) select id, user_name, first_name, last_name, photo_url from users where id=%s"
            db.engine.execute(q, id)
            db.session.commit()
        except Exception as e:
            print(e, file=sys.stderr)
    # return user for use        
    return Member.query.filter_by(id=id).first()

if __name__ == "__main__":
    flask_app.run(debug=False, host="msg", port=5000)