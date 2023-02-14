FROM golang:alpine AS builder
WORKDIR /tmp/pharmacy-auth
ADD . .
RUN apk add make && GOPATH=$(pwd)/cache make all

FROM alpine:3.17.2
RUN apk add bind-tools
COPY --from=0 /tmp/pharmacy-auth/build/pharmacy-auth /usr/bin/
COPY ./migrations /tmp/migrations
ENTRYPOINT ["pharmacy-auth"]