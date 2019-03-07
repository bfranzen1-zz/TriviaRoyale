# Build Docker Container
docker build -t bfranzen1/msg .

docker login
docker push bfranzen1/msg

# run mysql database
#ssh ec2-user@ec2-52-35-123-64.us-west-2.compute.amazonaws.com "
#docker rm -f msg;
#docker pull bfranzen1/msg &&
#docker run -d \
#--name msg \
#--network api \
#-e pw=\$MYSQL_ROOT_PASSWORD
#-e usr=root
#-e DBADDR=mysql:3306
#-e ADDR=msg
#bfranzen1/msg
#"