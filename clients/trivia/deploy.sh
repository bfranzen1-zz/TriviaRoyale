# Run build.sh
bash ./build.sh

# Push to docker hub
docker push jmatray/jmatray.me.trivia

# ssh into ec2 instance
ssh ec2-user@ec2-52-43-147-196.us-west-2.compute.amazonaws.com "
export TLSCERT=/etc/letsencrypt/live/client.bfranzen.me/fullchain.pem;
export TLSKEY=/etc/letsencrypt/live/client.bfranzen.me/privkey.pem;
docker rm -f jmatray.me.trivia;
docker pull jmatray/jmatray.me.trivia;
docker run -d --name jmatray.me.trivia -p 81:81 -v //etc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY jmatray/jmatray.me.trivia
"