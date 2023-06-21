targets := policy_enforcer orchestrator
sidecar_targets := rabbitmq-sidecar

prepare:
	go mod tidy
	go mod download

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/proto/*.proto

$(sidecar_targets): %: prepare proto
	cp Dockerfile go.mod go.sum ./cmd/$*/
	cp -r pkg ./cmd/$*/
	(trap 'rm -f ./cmd/$*/Dockerfile; rm -f ./cmd/$*/go.*; rm -rf ./cmd/$*/pkg' EXIT; \
	docker build --build-arg NAME='$*' -t $* ./cmd/$*/)

# orchestrator: rabbitmq-sidecar
# 	cp Dockerfile go.mod go.sum ./cmd/$*/
# 	cp -r pkg ./cmd/$*/
# 	(trap 'rm -f ./cmd/$*/Dockerfile; rm -f ./cmd/$*/go.*; rm -rf ./cmd/$*/pkg' EXIT; \
# 	docker build --build-arg NAME='$*' -t $* ./cmd/$*/)
$(targets): rabbitmq-sidecar
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	(trap 'rm -f ./cmd/$@/Dockerfile; rm -f ./cmd/$@/go.*; rm -rf ./cmd/$@/pkg' EXIT; \
	docker build --build-arg NAME='$@' -t $@ ./cmd/$@/)


all: $(targets) $(sidecar_targets) orchestrator

.PHONY: all $(targets) $(sidecar_targets) prepare proto orchestrator



# targets := anonymize query gateway agent orchestrator reasoner test rabbitmq-sidecar

# prepare:
# 	go mod tidy
# 	go mod download

# proto:
# 	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/proto/*.proto

# $(targets): %: prepare
# 	cp Dockerfile go.mod go.sum ./cmd/$*/
# 	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/proto/*.proto
# 	cp -r pkg ./cmd/$*/
# 	docker build --build-arg NAME='$*' -t $* ./cmd/$*/
# 	rm ./cmd/$*/Dockerfile
# 	rm ./cmd/$*/go.*
# 	rm -rf ./cmd/$*/pkg

# all: $(targets)

# .PHONY: all $(targets) prepare proto
