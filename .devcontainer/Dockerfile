FROM golang:1.21.6-bullseye

# Install Go tools
RUN apt-get update && apt-get install -y \
    git \
    unzip \
    curl \
    wget \
    protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
