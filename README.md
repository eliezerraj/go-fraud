# go-fraud
go-fraud

## Compile grpc proto

    protoc -I proto proto/fraud.proto --go_out=plugins=grpc:proto
