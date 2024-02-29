dynamos_path := /Users/jorrit/Documents/uva/DYNAMOS
config_path := $(dynamos_path)/configuration
k8s_service_files := $(config_path)/k8s_service_files
charts_path := $(dynamos_path)/charts
data_values_yaml := $(charts_path)/data-values.yaml
SHELL := /bin/zsh

create-rabbitmq-secret:
	@echo "Generating RabbitMQ password..."
	$(eval rabbit_pw=$(shell openssl rand -hex 16))
	$(eval rabbit_definitions_file=$(k8s_service_files)/definitions.json)

	@echo "Hashing password..."
	$(eval hashed_pw=$(shell docker run --rm rabbitmq:3-management rabbitmqctl hash_password ${rabbit_pw}))
	$(eval actual_hash1=$(shell echo "$${hashed_pw}" | cut -d $'\n' -f2))
	$(eval actual_hash=$(shell echo "$${hashed_pw}" | awk 'END{print $$NF}'))
	$(eval actual_hash3=$(shell echo "$${hashed_pw}" | tr -d '\r' | tr -d '\n' | awk 'END{print $$NF}'))

	@echo ${hashed_pw}
	@echo ${actual_hash}
	@echo ${actual_hash1}
	@echo ${actual_hash3}

	@echo "Replacing tokens..."
	cp $(k8s_service_files)/definitions_example.json $(rabbit_definitions_file)

	@echo "Performing token replacements..."
	@if [[ "$$OSTYPE" == "darwin"* ]]; then \
	    sed -i '' "s|%PWD%|${PWD}|g" $(data_values_yaml); \
	    sed -i '' "s|%PASSWORD%|${actual_hash}|g" $(rabbit_definitions_file); \
	else \
	    sed -i "s|%PWD%|${PWD}|g" $(data_values_yaml); \
	    sed -i "s|%PASSWORD%|${actual_hash}|g" $(rabbit_definitions_file); \
	fi

	@echo "RabbitMQ secret creation step completed."
