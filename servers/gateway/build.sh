# Build Go executable using linux
GOOS=linux go build

# Build Docker Containers
docker build -t bfranzen1/trivia.bfranzen.me .
docker build -t bfranzen1/tmysql ../db
# Delete pre-existing Go executable
go clean