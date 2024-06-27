targets := sql-query
# datasets := "/Users/jorrit/Documents/uva/thesis/DYNAMOS/configuration/datasets/"
datasets := "../configuration/datasets/"
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
dockerhub_account := dynamos1

.PHONY: proto $(targets)

proto:
	python3 -m grpc_tools.protoc -I../proto-files --python_out=./grpc_lib --grpc_python_out=./grpc_lib ../proto-files/*.proto

$(targets):
	cp Dockerfile ./$@
	cp -r grpc_lib ./$@
	cp -r $(datasets) ./$@

	(trap 'rm -f ./$@/Dockerfile; rm -rf ./$@/grpc_lib; rm -rf ./$@/datasets;' EXIT;  \
	docker build --build-arg NAME='$@' -t $(dockerhub_account)/$@:$(branch_name) -t $(dockerhub_account)/$@:latest ./$@/ )
	docker push $(dockerhub_account)/$@:$(branch_name)
	docker push $(dockerhub_account)/$@:latest


all: all $(targets) $(proto)
