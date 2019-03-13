# Build Go executable using linux
GOOS=linux go build

# Build Docker Container
docker build -t bfranzen1/trivia .

# Delete pre-existing Go executable
go clean

docker login
docker push bfranzen1/trivia