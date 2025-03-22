
# Default target
all: run

# Build
build:
	@echo "Build proto..."
	@protoc ./internal/core/proto/health/*.proto --go_out=. --go-grpc_out=.
	@protoc ./internal/core/proto/payment/*.proto --go_out=. --go-grpc_out=.
	@protoc ./internal/core/proto/pod/*.proto --go_out=. --go-grpc_out=.

# run
run:
	@echo "Run..."
	@go run ./cmd/main.go

.PHONY: all build run