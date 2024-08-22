SHELL := /bin/bash
targets := sql-query
datasets := "../configuration/datasets/"
branch_name := $(shell git rev-parse --abbrev-ref HEAD)
dockerhub_account := dynamos1

.PHONY: proto $(targets) dynamos

proto:
	@python3 -m grpc_tools.protoc -I../proto-files --python_out=./dynamos-python-lib/protofiles --grpc_python_out=./dynamos-python-lib/protofiles ../proto-files/*.proto

dynamos:
	(cd ./dynamos-python-lib && python3 setup.py bdist_wheel)

# Set PYTHONPATH to protofiles directory for includes to GRPC code to work
$(targets): proto
	cp Dockerfile ./$@

	# Build the wheel for the local package
	(cd ./dynamos-python-lib && python3 setup.py bdist_wheel)

	# Copy the wheel to the target directory
	cp ./dynamos-python-lib/dist/*.whl ./$@
	cp -r ./dynamos-python-lib/protofiles ./$@

	@trap 'rm -f ./$@/Dockerfile; rm -rf ./$@/*.whl; rm -rf ./$@/datasets; rm -rf ./$@/protofiles' EXIT;  \
	WHEEL_NAME=$$(ls ./$@/*.whl); \
	BASE=$$(basename $$WHEEL_NAME); \
	docker build --build-arg NAME=$@ -t $(dockerhub_account)/$@:$(branch_name) --build-arg WHEEL_NAME=$$BASE -t $(dockerhub_account)/$@:latest ./$@/
	docker push $(dockerhub_account)/$@:$(branch_name)
	docker push $(dockerhub_account)/$@:latest
all: all $(targets) $(proto)
