SHELL := /bin/bash
targets := mbt-adapter
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
dockerhub_account := dynamos1

# Set defaults for environment variables
ADAPTER_NAME ?= default_name
ADAPTER_URL ?= http://example.com
ADAPTER_TOKEN ?= default_token

.PHONY: all local $(targets) $(dev)

proto:
	protoc -I ../proto-files --ruby_out=./mbt-adapter/lib/messages ../proto-files/*.proto 
	protoc -I ./mbt-adapter/lib/dynamos/generic/proto --ruby_out=./mbt-adapter/lib/messages ./mbt-adapter/lib/dynamos/generic/proto/*.proto

$(targets): proto
	cp Dockerfile ./$@
	@trap 'rm -f ./$@/Dockerfile' EXIT; \
		docker build \
		--build-arg NAME=$@ \
		--build-arg ADAPTER_NAME=${ADAPTER_NAME} \
		--build-arg ADAPTER_URL=${ADAPTER_URL} \
		--build-arg ADAPTER_TOKEN=${ADAPTER_TOKEN} \
		-t $(dockerhub_account)/$@:$(branch_name) \
		./$@

	docker push $(dockerhub_account)/$@:$(branch_name)

dev: proto
	cp Dockerfile.dev ./mbt-adapter
	@trap 'rm -f ./mbt-adapter/lib/Dockerfile.dev' EXIT; \
		docker build \
		-f Dockerfile.dev \
		--build-arg NAME=adapter-dev \
		--build-arg ADAPTER_NAME=${ADAPTER_NAME} \
		--build-arg ADAPTER_URL=${ADAPTER_URL} \
		--build-arg ADAPTER_TOKEN=${ADAPTER_TOKEN} \
		-t adapter-dev:latest \
		./mbt-adapter

all: $(targets)


