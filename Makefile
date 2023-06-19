targets := anonymize query gateway agent orchestrator reasoner test

prepare:
	go mod tidy
	go mod download

$(targets): %: prepare
	cp Dockerfile go.mod go.sum ./cmd/$*_service/
	cp -r pkg ./cmd/$*_service/
	docker build --build-arg NAME='$*' -t $*_service ./cmd/$*_service/
	rm ./cmd/$*_service/Dockerfile
	rm ./cmd/$*_service/go.*
	rm -rf ./cmd/$*_service/pkg

all: $(targets)

.PHONY: all $(targets) prepare
