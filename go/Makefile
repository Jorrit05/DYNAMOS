targets := sidecar policy_enforcer orchestrator agent algorithm test anonymize api-gateway

prepare:
	go mod tidy
	go mod download

proto:
	protoc -I ../proto-files --go_out=./pkg/proto --go_opt=paths=source_relative --go-grpc_out=./pkg/proto --go-grpc_opt=paths=source_relative ../proto-files/*.proto

$(targets): prepare
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
	docker build --build-arg NAME='$@' -t jorrit05/dynamos-$@ ./cmd/$@/)
	docker push jorrit05/dynamos-$@


all: $(targets)

.PHONY: $(targets) prepare proto sidecar algorithm test anonymize 

