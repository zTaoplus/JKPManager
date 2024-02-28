FROM golang:1.22 AS builder


WORKDIR /app

COPY . .

RUN go build -o jpkmanager ./src


FROM ubuntu:22.04

WORKDIR /root/

COPY --from=builder /app/jpkmanager .


ENV JPK_EG_ENDPOINT=http://enterprise-gateway.jupyter.svc.cluster.local:8888
ENV JPK_MAX_PENDING_KERNELS=10
ENV JPK_REDIS_HOST=redis.tablegpt-test.svc.cluster.local
ENV JPK_ACTIVATION_INTERVAL=60

CMD ["./jpkmanager"]
