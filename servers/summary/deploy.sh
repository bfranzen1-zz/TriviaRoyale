# Build Go executable using linux
GOOS=linux go build

# Build Docker Container
docker build -t bfranzen1/summary .

# Delete pre-existing Go executable
go clean

docker login
docker push bfranzen1/summary

# pull and run Container from API VM
ssh ec2-user@ec2-52-35-123-64.us-west-2.compute.amazonaws.com "docker rm -f sum;
docker pull bfranzen1/summary &&
docker run -d \
--name sum \
--network api \
bfranzen1/summary
"