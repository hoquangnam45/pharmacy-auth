FROM golang:alpine AS builder
WORKDIR /tmp/pharmacy-auth
ADD . .
RUN apk add make && GOPATH=$(pwd)/cache make all

FROM alpine:3.17.2
RUN apk add bind-tools
COPY --from=builder /tmp/pharmacy-auth/build/pharmacy-auth /app
COPY ./migrations /app/migrations
WORKDIR /app
ENTRYPOINT ["./pharmacy-auth"]