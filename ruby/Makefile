SHELL := /bin/bash
targets := mbt-adapter
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
dockerhub_account := dynamos1

.PHONY: $(targets)


# Set defaults for environment variables
ADAPTER_NAME ?= default_name
ADAPTER_URL ?= http://example.com
ADAPTER_TOKEN ?= default_token

$(targets):
	cp Dockerfile ./$@
	@trap 'rm -f ./$@/Dockerfile' EXIT; \
		docker build \
		--build-arg NAME=$@ \
		--build-arg ADAPTER_NAME=my_adapter \
		--build-arg ADAPTER_URL=http://myurl.com \
		--build-arg ADAPTER_TOKEN=my_token \
		-t $(dockerhub_account)/$@:$(branch_name) \
		./$@

	docker push $(dockerhub_account)/$@:$(branch_name)
all: $(targets)
