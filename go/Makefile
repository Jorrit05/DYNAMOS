# Use Bash shell instead of default /bin/sh to support trap and extended features
SHELL := /bin/bash
# Replace slashes in the current Git branch name with dashes (e.g., issue/123 -> issue-123)
branch_name := $(shell git rev-parse --abbrev-ref HEAD | sed 's#/#-#g')
# List of SQL-based microservices to build and push
sql_microservices := sql-algorithm sql-anonymize sql-aggregate sql-test
# List of DYNAMOS core services
dynamos := sidecar policy-enforcer orchestrator agent api-gateway
# Docker Hub account for tagging and pushing images
dockerhub_account := poetoec

# Prepare Go project dependencies
# Clean and verify Go module dependencies
prepare:
	go mod tidy
	go mod download

# Compile Protocol Buffers into Go code with gRPC support (avoid inline comments, as they are printed when executing the Makefile)
# This target uses `protoc` (Protocol Buffers compiler) to generate both
# 	1. standard Go code from .proto definitions
# 	2. Go gRPC service stubs, which allow server/client communication via gRPC.
#
# The `-I ../proto-files` flag tells protoc where to find .proto definitions.
# 	--go_out: generates Go code for messages (structs, enums) and writes to pkg/proto.
# 	--go_opt=paths=source_relative: keeps file paths the same as proto structure.
# 	--go-grpc_out: generates gRPC service code.
# 	--go-grpc_opt=paths=source_relative: same path preservation as above.
#
# All generated files are saved in `pkg/proto` so they can be imported by Go modules.
proto:
	protoc -I ../proto-files \
		--go_out=./pkg/proto --go_opt=paths=source_relative \
		--go-grpc_out=./pkg/proto --go-grpc_opt=paths=source_relative \
		../proto-files/*.proto

# Pattern rule for building and pushing any service, allowing to reuse the same code for all services
# Design and general notes:
# 	- Inline comments are avoided because they are printed when executing the Makefile in the terminal
# 	- The $@ symbol represents the target name (e.g., sidecar from the dynamos variable list at the top of this file)
# 	- The build directory (./cmd/$@) is not set as a separate variable to avoid issues with make's automatic variable expansion
# This does the following:
# 	1. Copy the base Dockerfile and Go module files into the service's cmd directory
# 	2. Copy pkg directory into the service's cmd directory
# 	3. Use a subshell with a trap to clean up copied files after Docker build (comments must be outside to avoid errors):
# 		- The trap command must start and end with ( and ) on the first and last line respectively, otherwise will error: end of file unexpected (expecting ")")
# 		- This will build the Docker image with service name and tag with branch and 'latest'
# 	4. Push the Docker image to Docker Hub with both tags
define build_and_push
	cp Dockerfile go.mod go.sum ./cmd/$@
	cp -r pkg ./cmd/$@
	
	(trap 'rm -f ./cmd/$@/Dockerfile; \
			rm -f ./cmd/$@/go.*; \
			rm -rf ./cmd/$@/pkg' EXIT; \
		docker build --build-arg NAME='$@' \
			-t $(dockerhub_account)/$@:$(branch_name) \
			-t $(dockerhub_account)/$@:latest \
			./cmd/$@/)
	
	docker push $(dockerhub_account)/$@:$(branch_name)
	docker push $(dockerhub_account)/$@:latest
endef

# Rule to build and push Docker images for each DYNAMOS service
# Depends on 'prepare' and 'proto' steps
# Use the defined build_and_push rule to build and push each service
$(dynamos): prepare proto
	$(call build_and_push)

# Same as above, but for the SQL-based microservices
$(sql_microservices): prepare proto
	$(call build_and_push)

# Build and push all services
all: $(dynamos) $(sql_microservices)
# Convenience aliases to build specific groups of services
dynamos: $(dynamos)
sql_microservices: $(sql_microservices)

# Declare targets that are not associated with (physical) files.
# This prevents make from being misled by a file or directory that has the same name as the target.
# For example, if a file named "prepare" or "proto" exists, make might think it doesn't need to run the commands again.
# By declaring them .PHONY, make will always execute the associated commands when these targets are invoked.
.PHONY: all prepare proto $(dynamos) $(sql_microservices)
