# call build bash script
./build.sh

# deploy microservices
(cd ../messaging/ ; sh build.sh)

# build final project microservice
(cd ../trivia/ ; sh build.sh)

# push Container to Docker Hub
docker login
docker push bfranzen1/trivia.bfranzen.me
docker push bfranzen1/tmysql
docker push bfranzen1/trivia

# pull and run Container from API VM
# might be getting issues because of this >
        # docker system prune -a --volumes;
ssh ec2-user@ec2-52-26-94-110.us-west-2.compute.amazonaws.com " docker system prune -a --volumes;
docker network create api;
docker rm -f api;
docker rm -f mysql;
docker rm -f red;
docker rm -f rmq;
docker rm -f mgo;
docker run -d --name red --network api redis;
docker run -d --name rmq --network api -p 5672:5672 -p 15672:15672 rabbitmq:3-management;
export MYSQL_ROOT_PASSWORD=$(openssl rand -base64 18);
docker run -d --name mysql --network api \
-e MYSQL_ROOT_PASSWORD=\$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=api \
bfranzen1/tmysql;
export DSN=\"root:\$MYSQL_ROOT_PASSWORD@tcp(mysql:3306)/api\"; 
export SESSIONKEY=$(openssl rand -base64 18);
docker run -d --name mgo --network api -p 27017:27017 mongo;

sleep 20s;
docker exec mysql mysql -uroot -p\$MYSQL_ROOT_PASSWORD -e \"ALTER USER root IDENTIFIED WITH mysql_native_password BY '\$MYSQL_ROOT_PASSWORD';\"

docker rm -f msg;
docker pull bfranzen1/msg &&
docker run -d \
--name msg \
--network api \
-e pw=\$MYSQL_ROOT_PASSWORD \
-e usr=root \
-e DBADDR=mysql:3306 \
-e ADDR=msg \
-e mqHOST=rmq \
-e mqPORT=5672 \
-e rUSER='guest' \
-e rPW='guest' \
-e rmQueue='notify' \
bfranzen1/msg;

docker rm -f trivia;
docker pull bfranzen1/trivia && 
docker run -d \
--name trivia \
--network api \
-e RABBITMQ=rmq:5672 \
-e DSN=\$DSN \
-e MONGO_ADDR=mgo:27017 \
bfranzen1/trivia;

docker pull bfranzen1/trivia.bfranzen.me &&
docker run -d \
--name api \
--network api \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=/etc/letsencrypt/live/trivia.bfranzen.me/fullchain.pem \
-e TLSKEY=/etc/letsencrypt/live/trivia.bfranzen.me/privkey.pem \
-e SESSIONKEY=\$SESSIONKEY \
-e REDISADDR=red:6379 \
-e DSN=\$DSN \
-e MESSAGESADDR=msg:5000 \
-e RABBITMQ=rmq:5672 \
-e TRIVADDR=trivia:8000 \
bfranzen1/trivia.bfranzen.me;
"