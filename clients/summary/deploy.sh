# call build bash script
./build.sh

# push Container to Docker Hub
docker login
docker push bfranzen1/www.bfranzen.me

# pull and run Container from API VM
ssh ec2-user@ec2-54-71-20-80.us-west-2.compute.amazonaws.com "docker rm -f client;
docker pull bfranzen1/www.bfranzen.me &&
docker run -d \
--name client \
-p 80:80 \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=/etc/letsencrypt/live/www.bfranzen.me/fullchain.pem \
-e TLSKEY=/etc/letsencrypt/live/www.bfranzen.me/privkey.pem \
bfranzen1/www.bfranzen.me
"