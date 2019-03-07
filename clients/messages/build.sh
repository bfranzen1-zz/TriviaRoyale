# Build Go executable using linux
GOOS=linux go build

# Build Docker Container
docker build -t bfranzen1/signin .

# Delete pre-existing Go executable
go clean