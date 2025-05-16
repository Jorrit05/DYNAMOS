# Use Bash shell instead of the default /bin/sh to enable more advanced shell features
SHELL := /bin/bash
# List of service directories to process
targets := sql-query
# Path to datasets used by services (currently unused in this file)
datasets := "../configuration/datasets/"
# Replace slashes in the current Git branch name with dashes (e.g., issue/123 -> issue-123)
branch_name := $(shell git rev-parse --abbrev-ref HEAD | sed 's#/#-#g')
# Docker Hub account name used for tagging and pushing Docker images
dockerhub_account := poetoec

# Compile .proto files into Python and gRPC source code
# This target uses grpc_tools.protoc to generate both:
#   1. Python modules for Protobuf message classes (e.g., Foo_pb2.py)
#   2. Python modules for gRPC service stubs (e.g., Foo_pb2_grpc.py)
#
# -I ../proto-files       : Include path where .proto files are located
# --python_out            : Where to output the generated Protobuf message classes
# --grpc_python_out       : Where to output the generated gRPC service code
#
# The generated files are written into dynamos-python-lib/protofiles
# so they can be included as a module and bundled into the Python wheel (.whl)
# used in service Docker builds.
proto:
	@python3 -m grpc_tools.protoc \
		-I../proto-files \
		--python_out=./dynamos-python-lib/protofiles \
		--grpc_python_out=./dynamos-python-lib/protofiles \
		../proto-files/*.proto

# Build a wheel (.whl) package for the dynamos-python-lib project
dynamos:
	(cd ./dynamos-python-lib && python3 setup.py bdist_wheel)

# Pattern rule for building and pushing any service, allowing to reuse the same code for all services
# Also compiles GRPC proto files and builds the dynamos-python-lib wheel
# Set PYTHONPATH to protofiles directory for includes to GRPC code to work
# Design and general notes:
# 	- Inline comments are avoided because they are printed when executing the Makefile in the terminal
# 	- The $@ symbol represents the target name (e.g., sidecar from the dynamos variable list at the top of this file)
# 	- The build directory (./$@) is not set as a separate variable to avoid issues with make's automatic variable expansion
# This does the following:
# 	1. Copy the Dockerfile into the target's build context
# 	2. Build a Python wheel package of the dynamos-python-lib module for the local package
# 	3. Copy the built wheel into the target directory + Copy the generated GRPC Python code into the target directory
# 	3. Use a subshell with a trap to clean up copied files after Docker build (comments must be outside to avoid errors):
# 		- The trap command must start and end with ( and ) on the first and last line respectively, otherwise will error: end of file unexpected (expecting ")")
# 		- This first gets the filename of the copied wheel
# 		- Then the Docker image, passing service name and wheel file as build args, and tags with branch name and 'latest'
# 	4. Push the Docker image to Docker Hub with both tags
define build_and_push
	cp Dockerfile ./$@

	(cd ./dynamos-python-lib && python3 setup.py bdist_wheel)

	cp ./dynamos-python-lib/dist/*.whl ./$@
	cp -r ./dynamos-python-lib/protofiles ./$@

	(trap 'rm -f ./$@/Dockerfile; \
		      rm -rf ./$@/*.whl; \
		      rm -rf ./$@/datasets; \
		      rm -rf ./$@/protofiles' EXIT; \
		WHEEL_NAME=$$(ls ./$@/*.whl); \
		BASE=$$(basename $$WHEEL_NAME); \
		docker build \
			--build-arg NAME=$@ \
			--build-arg WHEEL_NAME=$$BASE \
			-t $(dockerhub_account)/$@:$(branch_name) \
			-t $(dockerhub_account)/$@:latest \
			./$@/)
	
	docker push $(dockerhub_account)/$@:$(branch_name)
	docker push $(dockerhub_account)/$@:latest
endef

# Build and push Docker images for each target
# Use the defined build_and_push rule to build and push each service
$(targets): proto
	$(call build_and_push)

# Convenience target to build all services and proto definitions
all: all $(targets) $(proto)

# Declare targets that are not associated with physical files
.PHONY: proto $(targets) dynamos
