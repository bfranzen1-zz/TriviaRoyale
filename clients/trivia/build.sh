# Build a go executable
GOOS=linux

#Build the docker container
docker build -t jmatray/jmatray.me.trivia .

# Delete Go executable
go clean