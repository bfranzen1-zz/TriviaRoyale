from datetime import datetime
import flask_sqlalchemy

db = flask_sqlalchemy.SQLAlchemy()

chann_mems = db.Table('chann_mems',
    db.Column('member_id', db.Integer, db.ForeignKey('member.id')),
    db.Column('channel_id', db.Integer, db.ForeignKey('channel.id'))
)

class Member(db.Model):
    __tablename__ = 'member'
    id = db.Column(db.Integer, primary_key=True)
    userName = db.Column(db.String(255), index=True)
    firstName = db.Column(db.String(64)) 
    lastName = db.Column(db.String(128))
    photoURL = db.Column(db.String(64))
    posts = db.relationship('Message', backref='creator', lazy='dynamic')
    channs = db.relationship('Channel', backref='creator', lazy='dynamic')
    def __repr__(self):
        return 'Member {}'.format(self.userName)
    
    def as_dict(self):
        d = {}
        for c in self.__table__.columns:
            d[c.name] = str(getattr(self, c.name))
        return d

class Message(db.Model):
    __tablename__ = 'message'
    id = db.Column(db.Integer, primary_key=True)
    body = db.Column(db.Text())
    created_at = db.Column(db.DateTime, index=True, default=datetime.utcnow)
    creator_id = db.Column(db.Integer, db.ForeignKey('member.id'))
    chann_id = db.Column(db.Integer, db.ForeignKey('channel.id'))
    edited_at = db.Column(db.DateTime)

    def __repr__(self):
        return '<Message {}>'.format(self.body)
    
    def as_dict(self):
        d = {}
        for c in self.__table__.columns:
            if c.name != "creator_id":
                d[c.name] = str(getattr(self, c.name))
        d["creator"] = self.creator.as_dict()
        return d


class Channel(db.Model):
    __tablename__ = 'channel'
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(255), index=True)
    description = db.Column(db.String(500))
    private = db.Column(db.Boolean, default=False)
    members = db.relationship('Member', secondary=chann_mems)
    created_at = db.Column(db.DateTime, index=True, default=datetime.utcnow)
    creator_id = db.Column(db.Integer, db.ForeignKey('member.id'))
    edited_at = db.Column(db.DateTime)

    def __repr__(self):
        return '<Channel {}>'.format(self.name)

    def as_dict(self):
        d = {}
        for c in self.__table__.columns:
            if c.name != "creator_id":
                d[c.name] = str(getattr(self, c.name))
        d["members"] = [m.as_dict() for m in self.members]
        d["creator"] = self.creator.as_dict()
        return d