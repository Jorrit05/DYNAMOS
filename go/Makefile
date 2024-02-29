targets := sidecar policy-enforcer orchestrator agent algorithm test anonymize api-gateway aggregate
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
api_targets := sidecar orchestrator api-gateway

prepare:
	go mod tidy
	go mod download

ifeq ($(shell uname),Linux)
proto:
	sudo docker run -v /home/ale/Dev/DYNAMOS:/data proto /bin/sh -c "cd /data/go && make proto"
else
proto:
	protoc -I ../proto-files --go_out=./pkg/proto --go_opt=paths=source_relative --go-grpc_out=./pkg/proto --go-grpc_opt=paths=source_relative ../proto-files/*.proto
endif

ifeq ($(shell uname),Linux)
$(targets): prepare
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
		(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
		sudo docker build --build-arg NAME='$@' -t jorrit05/dynamos-$@:$(branch_name) -t jorrit05/dynamos-$@:latest ./cmd/$@/)
		sudo docker push jorrit05/dynamos-$@:$(branch_name)
		sudo docker push jorrit05/dynamos-$@:latest
else
$(targets): prepare
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
	docker build --build-arg NAME='$@' -t jorrit05/dynamos-$@:$(branch_name) -t jorrit05/dynamos-$@:latest ./cmd/$@/)
	docker push jorrit05/dynamos-$@:$(branch_name)
	docker push jorrit05/dynamos-$@:latest
endif

all: $(targets)
api: $(api_targets)

.PHONY: $(targets) prepare proto sidecar algorithm test anonymize

