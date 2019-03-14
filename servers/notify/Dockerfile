# Notify microservice
FROM alpine 
RUN apk add --no-cache ca-certificates
COPY notify /notify
EXPOSE 443
ENTRYPOINT [ "/notify" ]