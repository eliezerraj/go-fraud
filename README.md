# go-fraud

    POC for test purposes.

## Compile grpc proto

    protoc -I proto proto/fraud.proto --go_out=plugins=grpc:proto
