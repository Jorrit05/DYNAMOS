targets := query

.PHONY: proto $(targets)

proto:
	python3 -m grpc_tools.protoc -I../proto-files --python_out=./query --grpc_python_out=./query ../proto-files/*.proto
$(targets):
	cp Dockerfile ./$@
	(trap 'rm -f ./$@/Dockerfile;' EXIT;  \
	docker build --build-arg NAME='$@' -t $@ ./$@/)


all: all $(targets) $(proto)
