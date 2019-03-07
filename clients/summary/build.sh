# Build Go executable using linux
GOOS=linux go build

# Build Docker Container
docker build -t bfranzen1/www.bfranzen.me .

# Delete pre-existing Go executable
go clean