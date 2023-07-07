targets := query
datasets := "/Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/configuration/datasets/"

.PHONY: proto $(targets)

proto:
	python3 -m grpc_tools.protoc -I../proto-files --python_out=./grpc_lib --grpc_python_out=./grpc_lib ../proto-files/*.proto
$(targets):
	cp Dockerfile ./$@
	cp -r grpc_lib ./$@
	cp -r $(datasets) ./$@

	(trap 'rm -f ./$@/Dockerfile; rm -rf ./$@/grpc_lib; rm -rf ./$@/datasets;' EXIT;  \
	docker build --build-arg NAME='$@' -t $@ ./$@/)


all: all $(targets) $(proto)
