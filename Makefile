targets := rabbitmq-sidecar policy_enforcer orchestrator


prepare:
	go mod tidy
	go mod download

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/proto/*.proto

$(targets): prepare proto
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
	docker build --build-arg NAME='$@' -t $@ ./cmd/$@/)


all: $(targets)

.PHONY: all $(targets) prepare proto rabbitmq-sidecar

