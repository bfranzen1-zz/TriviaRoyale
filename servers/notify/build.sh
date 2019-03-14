# Build Go executable using linux
go install
GOOS=linux go build

# Build Docker Container
docker build --no-cache -t bfranzen1/notify .

# Delete pre-existing Go executable
go clean