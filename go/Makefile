targets := sidecar policy-enforcer orchestrator agent algorithm test anonymize api-gateway aggregate
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
api_targets := sidecar orchestrator api-gateway
microservices := algorithm anonymize aggregate
agents := sidecar policy-enforcer orchestrator agent

prepare:
	go mod tidy
	go mod download

proto:
	protoc -I ../proto-files --go_out=./pkg/proto --go_opt=paths=source_relative --go-grpc_out=./pkg/proto --go-grpc_opt=paths=source_relative ../proto-files/*.proto

$(targets): prepare
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
	docker build --build-arg NAME='$@' -t jorrit05/dynamos-$@:$(branch_name) -t jorrit05/dynamos-$@:latest ./cmd/$@/)
	docker push jorrit05/dynamos-$@:$(branch_name)
	docker push jorrit05/dynamos-$@:latest

all: $(targets)
api: $(api_targets)
microservices: $(microservices)
agents: $(agents)

.PHONY: $(targets) prepare proto sidecar algorithm test anonymize

