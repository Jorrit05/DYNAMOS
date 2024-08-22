This document provides guidelines to installing all dependencies of DYNAMOS.

**NOTE**: This document mainly focuses on Linux based commands. Thus we expect Windows users to enable Windows Subsystem for Linux 2. If there are any special instructions related to other operating systems, it will clearly be highlighted.

# Prerequisites
## WSL2 (Windows)
There are multiple ways to enable this, whether that be through the CLI, the Windows store or some other magical way. Here is one example of a tutorial:
https://learn.microsoft.com/en-us/windows/wsl/install
## Docker Desktop
https://www.docker.com/products/docker-desktop/
Extra instructions:
1. Make sure to enable Kubernetes and provide sufficient resources to Docker when installed. We recommend at least 8GB of RAM and 4 CPU cores
2. **For Windows users**: Make sure to enable WSL integration in the settings menu. If the option is not there, you may need to edit a script within windows to enable the integration.
3.

# Installing

## Helm
Software to deploy all the DYNAMOS environment
https://helm.sh/docs/intro/install/
## Linkerd
https://linkerd.io/2.15/getting-started/
## kubectl
CLI tool to interact with Kubernetes:

On linux you can simply do:
```bash
sudo snap install kubectl --classic
```
https://kubernetes.io/docs/tasks/tools/
## Homebrew (for linux)

https://docs.brew.sh/Homebrew-on-Linux

## k9s (optional)
Preqeqs: Homebrew
Useful CLI tool to view all currently running kuberentes containers.
https://k9scli.io/topics/install/


# Secondary Installations

## Linkerd
### Prepare cluster
```bash
linkerd install --crds | kubectl apply -f -
linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -
linkerd check

linkerd jaeger install | kubectl apply -f -
# Maybe linerkd wiz install | kubectl apply -f -
```

### RabbitqMQ
```bash

# Hash password
docker run --rm rabbitmq3-management rabbitmqctl hash_password $pw

# Install prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade -i -f "${coreChart}/prometheus-values.yaml" prometheus prometheus-community/prometheus

# Install nginx
helm install -f "${coreChart}/ingress-values.yaml" nginx oci://gher.io/nginxinc/charts/
```


## Update bashrc or zshrc
```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
#export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
export DYNAMOS_ROOT="<CHANGE TO ABS PATH TO DYNAMOS REPO>"
export AMQ_PASSWORD_FILE="${DYNAMOS_ROOT}/go/cmd/user/rabbit_secret"
export VITE_APP_CLIENT_ID="f07e1dcb-9675-43d5-b7b7-41a80414e100"
export VITE_APP_AUTHORITY="https://login.microsoftonline.com/db02e38a-3a2b-4c8f-b937-b835f156198b"
export AMQ_USER="normal_user"
export LOCAL_DEV="true"
export PYTHONPATH="${DYNAMOS_ROOT}/python/grpc_lib"
export OC_AGENT_HOST="localhost:32002"

deploy_core() {
  coreChart="${DYNAMOS_ROOT}/charts/core"
  helm upgrade -i -f "${coreChart}/values.yaml" core ${DYNAMOS_ROOT}/charts/core --set hostPath="${DYNAMOS_ROOT}"
}

deploy_prometheus() {
  coreChart="${DYNAMOS_ROOT}/charts/core"
  helm upgrade -i -f "${coreChart}/prometheus-values.yaml" prometheus prometheus-community/prometheus
}

deploy_orchestrator() {
orchestratorChart="${DYNAMOS_ROOT}/charts/orchestrator/values.yaml"
  helm upgrade -i -f "${orchestratorChart}" orchestrator ${DYNAMOS_ROOT}/charts/orchestrator --set hostPath="${DYNAMOS_ROOT}"
}

deploy_api_gateway() {
  apiGatewayChart="${DYNAMOS_ROOT}/charts/api-gateway/values.yaml"
  helm upgrade -i -f "${apiGatewayChart}" api-gateway ${DYNAMOS_ROOT}/charts/api-gateway --set hostPath="${DYNAMOS_ROOT}"
}

deploy_agent() {
  agentChart="${DYNAMOS_ROOT}/charts/agents/values.yaml"
  helm upgrade -i -f "${agentChart}" agent ${DYNAMOS_ROOT}/charts/agents
}

deploy_surf() {
  surfChart="${DYNAMOS_ROOT}/charts/thirdparty/values.yaml"
  helm upgrade -i -f "${surfChart}" surf ${DYNAMOS_ROOT}/charts/thirdparty
}

deploy_ingress() {
  coreChart="${DYNAMOS_ROOT}/charts/core"
  helm install -f "${coreChart}/ingress-values.yaml" nginx ingress-nginx/ingress-nginx -n ingress
}

deploy_namespaces() {
  namespaceChart="${DYNAMOS_ROOT}/charts/namespaces"
  helm upgrade -i -f "${namespaceChart}/values.yaml" namespaces ${DYNAMOS_ROOT}/charts/namespaces --set hostPath="${DYNAMOS_ROOT}"
}

deploy_all() {
  deploy_namespaces
  deploy_agent
  deploy_core
  deploy_orchestrator
  deploy_api_gateway
  deploy_surf
  deploy_prometheus
}

delete_jobs() {
  kubectl get pods -A | grep 'jorrit-stutterheim' | awk '{split($2,a,"-"); print $1" "a[1]"-"a[2]"-"a[3]}' | xargs -n2 bash -c 'kubectl delete job $1 -n $0'
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/UVA/queueInfo/jorrit-stutterheim- --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/SURF/queueInfo/jorrit-stutterheim- --prefix
}

deploy_all(){
  deploy_orchestrator
  deploy_agent
  deploy_surf
  deploy_api_gateway
  deploy_core
}

deploy_addons(){
  deploy_orchestrator
  deploy_agent
  deploy_surf
  deploy_api_gateway
}

uninstall_all(){
  helm uninstall orchestrator
  helm uninstall surf
  helm uninstall agent
  helm uninstall api-gateway
  helm uninstall core
}

uninstall_addons(){
  helm uninstall orchestrator
  helm uninstall surf
  helm uninstall agent
  helm uninstall api-gateway
}

uninstall_orch(){
  helm uninstall orchestrator
}

watch_pods(){
  watch kubectl get pods -A
}

restart_core() {
  kubectl rollout restart deployment/rabbitmq -n core
}

redeploy_api_gateway(){
  helm uninstall api-gateway
  deploy_api_gateway
}

uninstall_api() {
  helm uninstall api-gateway
  helm uninstall orchestrator
}

deploy_api() {
  deploy_orchestrator
  deploy_api_gateway
}
```

## Deploy namespaces
cd into DYNAMOS project
```bash
cd /charts/namespaces
helm install .
```

## Delete the rabbit secrets
```bash
kubectl delete secret rabbit -n orchestrator
kubectl delete secret rabbit -n uva
kubectl delete secret rabbit -n vu
kubectl delete secret rabbit -n surf
kubectl delete secret rabbit -n api-gateway
```
## Configure Rabbit PVC
```bash
cd configuration
./fill-rabbit-pvc.sh
```

## Create definitions file
```bash
cp configuration/k8s_service_files/definitions_example.json configuration/k8s_service_files/definitions.json
```
Update the docker password in the definitions file
```bash
Send an email to create Jorrit to get the password
```

## Update hostfile
```
sudo vim /etc/hosts
127.0.0.1 api-gateway.api-gateway.svc.cluster.local
```

## Add ingress
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
```

## Deploy Ingress
```bash
deploy_ingress
```
