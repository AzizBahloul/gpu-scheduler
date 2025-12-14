#!/bin/bash

set -e

echo "Generating protobuf code..."

# Create output directory
mkdir -p pkg/api/grpc/generated

# Generate Go code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/grpc/scheduler.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/grpc/agent.proto

echo "Protobuf code generation complete!"
