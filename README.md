# go-fraud

    POC for test purposes.

## Compile grpc proto

    cd /mnt/c/Eliezer/workspace/github.com/go-payment/internal
    protoc -I proto proto/fraud.proto --go_out=plugins=grpc:proto

    $
    protoc --go_out=. ./internal/core/proto/pod/*.proto

## Handle certs tls

    Convert the certs to base64 sercer.cert and server.key

    /mnt/c/Eliezer/workspace/github.com/go-fraud/certs$ base64 -w 0 server.crt >> serverB64.crt
    /mnt/c/Eliezer/workspace/github.com/go-fraud/certs$ base64 -w 0 server.key >> serverB64.key
    /mnt/c/Eliezer/workspace/github.com/go-fraud/certs$ base64 -w 0 ca.crt > caB64.crt

    $/mnt/c/Eliezer/workspace/github.com/go-fraud/internal/proto$
    grpcurl -cacert=/mnt/c/Eliezer/workspace/github.com/go-fraud/certs/server.crt --proto ./fraud.proto localhost:50052 fraud.FraudService/
    
    