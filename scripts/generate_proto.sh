#!/bin/bash

# Check if protoc is installed
if ! command -v protoc >/dev/null 2>&1; then
  echo 'Error: protoc is not installed. Please install Protocol Buffers compiler.' >&2
  exit 1
fi

# Check if Go plugins for protoc are installed
if ! command -v protoc-gen-go >/dev/null 2>&1 || ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo 'Installing Go plugins for protoc...'
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
  export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Compile all .proto files in-place
find pkg/proto -name "*.proto" | while read -r proto_file; do
  echo "Generating code for $proto_file..."
  protoc -I=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    "$proto_file"
done

echo "All proto files generated successfully"