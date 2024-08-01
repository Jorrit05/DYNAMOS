# Installation Guide

TODO:
  - Add improved / automated way to fetch rabbitmq password 
  - Add description for PVC part
  - 

This document provides guidelines to installing all dependencies of DYNAMOS.

**NOTE**: This document mainly focuses on Linux Debian based commands. Thus we expect Windows users to enable Windows Subsystem for Linux 2. If there are any special instructions related to other operating systems, it will clearly be highlighted.

# Prerequisite software tools
In order to setup DYNAMOS, we need to install a few tools as a foundation. This section covers such tools.

## WSL2 (for Windows)
Linux based operating systems must be used for DYNAMOS. If Windows must be used, we suggest enabling the Windows Subsystem for Linux (WSL2).

There are multiple ways to enable this, whether that be through the CLI, the Windows store or some other magical way. Here is one example of a tutorial:

https://learn.microsoft.com/en-us/windows/wsl/install

## Docker Desktop
For local development we utilise Docker Desktop, as upon installation we get Kubernetes for free in an easily manageable interface. 

**NOTE**: In our development cycle we build and upload "finalized" images to docker hub, thus having a docker hub account may be useful if you intend to further develop services in DYNAMOS.

https://www.docker.com/products/docker-desktop/

Extra instructions: 
1. Make sure to enable Kubernetes and provide sufficient resources to Docker when installed. We recommend at least 8GB of RAM and 4 CPU cores.
2. **For Windows users**: Make sure to enable WSL integration in the settings menu. If the option is not there, you may need to edit a script within windows to enable the integration. If you are encountering issues with enabling WSL on Docker, use this blog as a reference: https://docs.docker.com/desktop/wsl/

## Homebrew (optional)
A package manager that may ease the setup process. This is automatically installed on most MacOS systems, however it requires some setup for Linux. 

Homebrew for Linux:

https://docs.brew.sh/Homebrew-on-Linux


Homebrew is not available for Windows, however it is available for WSL. 


# Installing
Now that we have our environment setup, we can start installing the required software to deploy DYNAMOS.

## kubectl
https://kubernetes.io/docs/tasks/tools/

Kubernetes CLI tool. 

Now that docker desktop has installed kubernetes for us, we need to be able to communicate with it, kubectl is a useful CLI tool for this purpose.

Homebrew:
```bash
brew install kubectl
```

Snap:
```bash
sudo snap install kubectl --classic
```

## Helm CLI
https://helm.sh/docs/intro/install/

Tool that is responsible for deploying the microservices to kubernetes based off of helm charts.

Helm can be installed via a script provided by helm or through a package manager

**Brew**:
```bash
brew install helm
```

**Apt**:
```bash
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

## Linkerd
https://linkerd.io/2.15/getting-started/

**Prereqs**: kubectl

Linkerd is a service mesh for Kubernetes, we use it to install Jaeger (https://www.jaegertracing.io/) a distributed tracing platform onto our cluster.

```bash
# Install CLI
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge | sh

# Add Linkerd to PATH
export PATH=$HOME/.linkerd2/bin:$PATH

# Install Linkerd on cluster
linkerd install --crds | kubectl apply -f -
```

## k9s (optional) 
https://k9scli.io/topics/install/

**Preqeqs**: Homebrew

If you don't want to constantly be using `kubectl` to check the status of your cluster, k9s can be used to visualize all the containers runnings within a namespace and within clusters.

HINT: When running k9s, it will initially load the default namespaces, press 0 to show all namespaces, which your containers are likely going to show up in.

```bash
 brew install derailed/k9s/k9s
```

# System Configuration
Now that we have the required software installed, we can start configuring our system to deploy DYNAMOS to kubernetes.

## Linkerd
Use Linkerd to install jaeger onto the cluster. Other services can be installed using Linkerd, such as security and reliability related software.
```bash
# You probably already ran the following command when installing
linkerd install --crds | kubectl apply -f -

linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -
linkerd check

# Install Jaeger onto the cluster for observability
linkerd jaeger install | kubectl apply -f -

# Optionally install for insight dashboard - not currently in use 
# linkerd wiz install | kubectl apply -f - 
```

## Add DYNAMOS env vars and helper functions to shell
To make the deployment process easier, we have prepared a set of environment variables and methods that can be added to your shell rc file. These are usually the `bashrc` or `zshrc` files. Alternatively, the below commands can be added to an additional file, and included in the shell file.

NOTE: A few steps (that we highlight in the latter part of the guide) are required before you can use the methods provided below.

```bash
## DYNAMOS Configs

# The directory where DYNAMOS repo is cloned
export DYNAMOS_ROOT="<CHANGE TO ABSOLUTE PATH TO DYNAMOS REPO>"
# Helm chart location for the core chart (used in multiple deployments)
export coreChart="${DYNAMOS_ROOT}/charts/core"

######################################
## Independant services deployment ##
#####################################

########################
### Generic services ###
########################

# Deploy the namespaces in the cluster
#   Important to use upon first installing DYNAMOS, without it k8 cannot be used
deploy_namespaces() {
  namespaceChart="${DYNAMOS_ROOT}/charts/namespaces"
  helm upgrade -i -f "${namespaceChart}/values.yaml" namespaces ${DYNAMOS_ROOT}/charts/namespaces --set hostPath="${DYNAMOS_ROOT}"
}

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
  helm upgrade -i -f "${orchestratorChart}" orchestrator ${DYNAMOS_ROOT}/charts/orchestrator --set hostPath="${DYNAMOS_ROOT}"
}

# Deploy the api-gateway to the cluster
#   Responsible of accepting any requests from the public, forwards the requests into the cluster
deploy_api_gateway() {
  apiGatewayChart="${DYNAMOS_ROOT}/charts/api-gateway/values.yaml"
  helm upgrade -i -f "${apiGatewayChart}" api-gateway ${DYNAMOS_ROOT}/charts/api-gateway --set hostPath="${DYNAMOS_ROOT}"
}

#####################################
### Specific to the AMDEX usecase ###
#####################################

# Deploy the agents (UVA, VU, etc) to the cluster 
deploy_agent() {
  agentChart="${DYNAMOS_ROOT}/charts/agents/values.yaml"
  helm upgrade -i -f "${agentChart}" agent ${DYNAMOS_ROOT}/charts/agents
}

# Deploy trusted third party "surf" to the cluster
deploy_surf() {
  surfChart="${DYNAMOS_ROOT}/charts/thirdparty/values.yaml"
  helm upgrade -i -f "${surfChart}" surf ${DYNAMOS_ROOT}/charts/thirdparty
}

##############################
## Bulk deployment commands ##
##############################

# Deploy all services, this can be used to deploy all services at one go
#   This runs the risk of having race conditions on the first attempt, however the system
#   should be able to recover automatically
deploy_all() {
  deploy_namespaces
  deploy_core
  deploy_orchestrator
  deploy_api_gateway
  deploy_agent
  deploy_surf
  deploy_prometheus
}

# Remove all services running in the cluster.
#   The namespaces are not removed since that is a layer above the other services
#   Any service can be manually removed by using `helm uninstall <name of service>` 
uninstall_all(){
  helm uninstall orchestrator
  helm uninstall surf
  helm uninstall agent
  helm uninstall api-gateway
  helm uninstall core
}

###################
## Misc Commands ##
###################

# Delete currently running jobs, this may be useful for stale jobs that remain alive for some reason.
delete_jobs() {
  kubectl get pods -A | grep 'jorrit-stutterheim' | awk '{split($2,a,"-"); print $1" "a[1]"-"a[2]"-"a[3]}' | xargs -n2 bash -c 'kubectl delete job $1 -n $0'
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/UVA/queueInfo/jorrit-stutterheim- --prefix
  etcdctl --endpoints=http://localhost:30005 del /agents/jobs/SURF/queueInfo/jorrit-stutterheim- --prefix
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

```

**Important**: Remeber to source your shell file after inserting the above into it.
```bash
  source ~/.bashrc
```
or
```bash
  source ~/.zshrc 
```
(or whatever shell rc you use)

## Deploy namespaces
After sourcing your shell rc file, register all namespaces to your cluster with:
```bash
  deploy_namespaces
```
This is required for the next step, where we register the RabbitMQ secret on all namespaces.

## Install and deploy Prometheus and Nginx with Helm 

```bash
# Install and deploy prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
deploy_prometheus

# Install and deploy nginx
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install -f "${coreChart}/ingress-values.yaml" nginx oci://gher.io/nginxinc/charts/
deploy_ingress
```

## Create password for RabbitMQ user
RabbitMQ requires a password to run properly, we do so via the 
```bash
# Create a password for a rabbit user
pw=$(openssl rand -base64 12)

# Add password to all namespaces
kubectl create secret generic rabbit --from-literal=password=${pw} -n orchestrator
kubectl create secret generic rabbit --from-literal=password=${pw} -n api_gateway
kubectl create secret generic rabbit --from-literal=password=${pw} -n uva
kubectl create secret generic rabbit --from-literal=password=${pw} -n vu
kubectl create secret generic rabbit --from-literal=password=${pw} -n surf
#  If there are any new namespaces, add them here

# Hash password 
docker run --rm  rabbitmq:3-management rabbitmqctl hash_password $pw
```
**Important**: Save the password we are going to use it in the next step!

## Create definitions file
From the root directory of DYNAMOS, create a copy of the definitions file like so:
```bash
cp configuration/k8s_service_files/definitions_example.json configuration/k8s_service_files/definitions.json
```
Update the docker password in the definitions file with the $pw we created in the previous step

## Configure Rabbit PVC
```bash
cd configuration
./fill-rabbit-pvc.sh
```

## Update hostfile
To be able to access DYNAMOS from your local machine, you'll need to add the `api-gateway` service to your hosts file.
To do this on Linux, use your favourite text editor with root access on the file `/etc/hosts`, like so:
```bash
sudo vim /etc/hosts
```
Now add the following to hosts file:
```bash
127.0.0.1 api-gateway.api-gateway.svc.cluster.local
```
Since the API gateway is the only public facing service, it is the only entry required in the hosts file. If any additional services are added, they should also be added here with a similiar pattern.

Note that this is super useful when trying to test DYNAMOS locally using tools such as `curl` or `postman`.

## Example Request
To make sure we installed everything properly, let's use the AMDeX use case as an example.
Firstly, make sure you've deployed everything, you can do that using `deploy_all()` in your CLI, or individually deploying each service and checking the status on k9s or using the `watch_pods()` method.

Let's setup the request.

The URL should be:
```
http://api-gateway.api-gateway.svc.cluster.local:32093/api/v1/requestApproval
```
as a **POST** request, with the following body with **JSON** encoding:
```json
{
    "type": "sqlDataRequest",
    "user": {
        "id": "12324",
        "userName": "jorrit.stutterheim@cloudnation.nl"
    },
    "dataProviders": ["VU","UVA","RUG"],
    "data_request": {
        "type": "sqlDataRequest",
        "query" : "SELECT * FROM Personen p JOIN Aanstellingen s LIMIT 1000",
        "algorithm" : "average",
        "options" : {
            "graph" : false,
            "aggregate": false 
        },
        "requestMetadata": {}   
    }
}
```

## Delete the rabbit secrets
```bash
kubectl delete secret rabbit -n orchestrator
kubectl delete secret rabbit -n uva
kubectl delete secret rabbit -n vu
kubectl delete secret rabbit -n surf
kubectl delete secret rabbit -n api-gateway
```
