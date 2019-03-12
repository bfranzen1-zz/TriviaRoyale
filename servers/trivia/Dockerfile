# Trivia Microservice
FROM alpine 
RUN apk add --no-cache ca-certificates
COPY trivia /trivia
EXPOSE 8000
ENTRYPOINT [ "/trivia" ]