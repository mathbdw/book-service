# syntax=docker/dockerfile:1

ARG GITHUB_PATH=github.com/mathbdw/book

FROM golang:1.24-alpine AS builder

ARG GITHUB_PATH

WORKDIR /home/${GITHUB_PATH}

RUN apk add --no-cache --update \
    make \
    git \
    protoc \
    protobuf \
    protobuf-dev \
    curl \
    && rm -rf /var/cache/apk/*

COPY Makefile .
RUN make deps-go
COPY . .
RUN make build-go

FROM alpine:latest as server
RUN apk --no-cache add ca-certificates
WORKDIR /root/

ARG GITHUB_PATH

COPY --from=builder /home/${GITHUB_PATH}/bin/book-service .
COPY --from=builder /home/${GITHUB_PATH}/config.yml .
COPY --from=builder /home/${GITHUB_PATH}/migrations/ ./migrations

RUN chown root:root book-service

EXPOSE 50051
EXPOSE 7001
EXPOSE 9100

CMD ["./book-service"]