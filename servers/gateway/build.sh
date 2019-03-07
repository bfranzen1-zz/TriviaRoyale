# Build Go executable using linux
GOOS=linux go build

# Build Docker Containers
docker build -t bfranzen1/api.bfranzen.me .
docker build -t bfranzen1/mysql ../db
# Delete pre-existing Go executable
go clean