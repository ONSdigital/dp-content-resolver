FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-content-resolver .

ENTRYPOINT ./dp-content-resolver
