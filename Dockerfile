# docker build -t go-grpc-service-server . -f dockerfile-server
FROM golang:1.21 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-fraud -ldflags '-linkmode external -w -extldflags "-static"'

WORKDIR /bin
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.6 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

FROM alpine

WORKDIR /bin
COPY --from=builder /bin/grpc_health_probe .

WORKDIR /app
COPY --from=builder /app/cmd/go-fraud .
CMD ["/app/go-fraud"]