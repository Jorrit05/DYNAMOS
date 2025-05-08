#!/bin/bash

## DYNAMOS Configs

# The directory where DYNAMOS repo is cloned
export DYNAMOS_ROOT="${HOME}/DYNAMOS"
# Helm chart location for the core chart (used in multiple deployments)
export coreChart="${DYNAMOS_ROOT}/charts/core"

######################################
## Independant services deployment ##
#####################################

########################
### Generic services ###
########################

# Deploy the nginx ingress to the cluster
#   This only needs to be done once
deploy_ingress() {
  helm install -f "${coreChart}/ingress-values.yaml" nginx ingress-nginx/ingress-nginx -n ingress
}

# Deploy the core services to the cluster
deploy_core() {
  helm upgrade -i -f "${coreChart}/values.yaml" core ${DYNAMOS_ROOT}/charts/core --set hostPath="${DYNAMOS_ROOT}"
}

# Deploy prometheus for metric collection
#   Only needs to be deployed once
deploy_prometheus() {
  helm upgrade -i -f "${coreChart}/prometheus-values.yaml" prometheus prometheus-community/prometheus
}

# Deploy the orchestator service to the cluster
#   Responsible for managing requests within DYNAMOS
deploy_orchestrator() {
  orchestratorChart="${DYNAMOS_ROOT}/charts/orchestrator/values.yaml"
  helm upgrade -i -f "${orchestratorChart}" orchestrator ${DYNAMOS_ROOT}/charts/orchestrator
}

# Deploy the api-gateway to the cluster
#   Responsible of accepting any requests from the public, forwards the requests into the cluster
deploy_api_gateway() {
  apiGatewayChart="${DYNAMOS_ROOT}/charts/api-gateway/values.yaml"
  helm upgrade -i -f "${apiGatewayChart}" api-gateway ${DYNAMOS_ROOT}/charts/api-gateway
}

#####################################
### Specific to the AMDEX usecase ###
#####################################

# Deploy the agents (UVA, VU, etc) to the cluster
deploy_agent() {
  agentChart="${DYNAMOS_ROOT}/charts/agents/values.yaml"
  helm upgrade -i -f "${agentChart}" agents ${DYNAMOS_ROOT}/charts/agents
}

# Deploy trusted third party "surf" to the cluster
deploy_surf() {
  surfChart="${DYNAMOS_ROOT}/charts/thirdparty/values.yaml"
  helm upgrade -i -f "${surfChart}" surf ${DYNAMOS_ROOT}/charts/thirdparty
}

# Deploy all for AMDEX (agents and third party: surf)
deploy_all_amdex() {
  deploy_agent
  deploy_surf
}
# Uninstall all for AMDEX (agents and third party: surf)
uninstall_all_amdex(){
  helm uninstall surf
  helm uninstall agents
}

##############################
## Bulk deployment commands ##
##############################

# Deploy all services, this can be used to deploy all services at one go
#   This runs the risk of having race conditions on the first attempt, however the system
#   should be able to recover automatically
deploy_all() {
  deploy_core
  deploy_orchestrator
  deploy_api_gateway
  deploy_agent
  deploy_surf
}

# Remove all services running in the cluster.
#   The namespaces are not removed since that is a layer above the other services
#   Any service can be manually removed by using `helm uninstall <name of service>`
uninstall_all(){
  helm uninstall orchestrator
  helm uninstall surf
  helm uninstall agents
  helm uninstall api-gateway
  helm uninstall core
}

# Deploy or remove all services excepting import core functionality.
deploy_addons() {
  deploy_orchestrator
  deploy_api_gateway
  deploy_agent
  deploy_surf
}

uninstall_addons() {
  helm uninstall orchestrator
  helm uninstall surf
  helm uninstall agents
  helm uninstall api-gateway
}

###################
## Misc Commands ##
###################

# Delete currently running jobs, this may be useful for stale jobs that remain alive for some reason.
delete_jobs() {
  kubectl get pods -A | grep 'jorrit-stutterheim' | awk '{split($2,a,"-"); print $1" "a[1]"-"a[2]"-"a[3]}' | xargs -n2 bash -c 'kubectl delete job $1 -n $0'
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/UVA/queueInfo/jorrit-stutterheim- --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/SURF/queueInfo/jorrit-stutterheim- --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/VU/queueInfo/jorrit-stutterheim- --prefix
}

# Delete jobs that may not be automatically deleted
delete_jobs_other() {
  kubectl get pods -A | grep 'jorrit-stutterheim' | awk '{split($2,a,"-"); print $1" "a[1]"-"a[2]"-"a[3]}' | xargs -n2 bash -c 'kubectl delete job $1 -n $0'
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/UVA/jorrit.stutterheim --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/SURF/jorrit.stutterheim --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/VU/jorrit.stutterheim --prefix
}

# Provides an overview of all running pods
#   Decent but basic alternative to using k9s
watch_pods(){
  watch kubectl get pods -A
}

# Restart RabbitMQ service
#   Could be useful if an error with the queue occurs during development
restart_core() {
  kubectl rollout restart deployment/rabbitmq -n core
}

# Redeploy all DYNAMOS components structurally. This ensures core is uninstalled first, then orchestrator, etc.
# If it is not done in this sequence, occassionally it causes some issues, such as pods not running, requests timing out, etc.
redeploy_structurally() {
  # Uninstall all and wait a while before it is all removed
  uninstall_all
  echo "Waiting for 1 minute..."
  sleep 60

  # Deploy core and wait for 15 seconds so orchestrator is running after core is done
  deploy_core
  echo "Waiting for 15 seconds..."
  sleep 15

  # Deploy orchestrator and wait a short time before deploying the rest
  deploy_orchestrator
  echo "Waiting for 10 seconds..."
  sleep 10

  # Deploy other components (sequence in which it is running does not matter here)
  deploy_agents
  deploy_api_gateway
}