FROM golang:1.22 AS builder


WORKDIR /app

COPY . .

RUN go build -o jkpmanager ./src


FROM ubuntu:22.04

WORKDIR /root/

COPY --from=builder /app/jkpmanager .


CMD ["./jkpmanager"]
