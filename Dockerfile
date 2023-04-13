FROM golang:1.18-alpine3.17 AS builder
WORKDIR /tmp/pharmacy-auth
ADD . .
RUN apk add make && GOPATH=$(pwd)/cache make build

FROM alpine:3.17.2 AS app
RUN apk add bind-tools
COPY --from=builder /tmp/pharmacy-auth/bin/pharmacy_auth /app/pharmacy_auth
COPY ./migrations /app/migrations/
WORKDIR /app
ENTRYPOINT ["./pharmacy_auth", "-conf", "./migrations"]