# go-fraud

    POC for test purposes.

## Compile grpc proto

    cd /mnt/c/Eliezer/workspace/github.com/go-payment/internal
    protoc -I proto proto/fraud.proto --go_out=plugins=grpc:proto
