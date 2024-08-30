# DYNAMOS

DYNAMOS: Dynamically Adaptive Microservice-based OS. A Middleware for Data Exchange Systems.

This repository contains the code for the DYNAMOS proof of concept. Installation instructions and more details can be found in the thesis.

More info, a demo video, and the thesis can be found on the website of [C. de Laat](https://delaat.net/sc/sc23/index.html)
The video unfortunately does not provide a voice-over yet.

The Fabric-CloudLab profile can be cloned and adapted from the [dynamos-cluster](https://github.com/Jorrit05/DYNAMOS-cluster) GitHub page.

# Table of Contents

- [DYNAMOS](#dynamos)
- [Table of Contents](#table-of-contents)
- [Installation Guide](#installation-guide)
- [Prerequisite software tools](#prerequisite-software-tools)
  - [1. WSL2 (for Windows)](#1-wsl2-for-windows)
  - [2. Docker Desktop](#2-docker-desktop)
  - [3. Homebrew (optional)](#3-homebrew-optional)
- [Installing](#installing)
  - [4. kubectl](#4-kubectl)
  - [5. Helm CLI](#5-helm-cli)
  - [6. Linkerd](#6-linkerd)
  - [7. etcdctl (optional)](#7-etcdctl-optional)
  - [8. k9s (optional)](#8-k9s-optional)
- [System Configuration](#system-configuration)
  - [9. Install script](#9-install-script)
  - [10. RabbitMQ password process](#10-rabbitmq-password-process)
    - [10.1. Configure Rabbit PVC](#101-configure-rabbit-pvc)
  - [11. Ingress](#11-ingress)
    - [11.1. Update hostfile](#111-update-hostfile)
- [Bashrc shortcuts](#bashrc-shortcuts)
  - [12. Add DYNAMOS env vars and helper functions to shell](#12-add-dynamos-env-vars-and-helper-functions-to-shell)
- [Example](#example)
  - [13. Example Request](#13-example-request)
- [Troubleshooting](#troubleshooting)
  - [14. Services crash because the connection to RabbitMQ does not work](#14-services-crash-because-the-connection-to-rabbitmq-does-not-work)


# Installation Guide

This document provides guidelines to installing all dependencies of DYNAMOS.

**NOTE**: This document mainly focuses on Linux Debian based commands. Thus we expect Windows users to enable Windows Subsystem for Linux 2. If there are any special instructions related to other operating systems, it will clearly be highlighted.

# Prerequisite software tools
In order to setup DYNAMOS, we need to install a few tools as a foundation. This section covers such tools.

##  1. <a name='WSL2forWindows'></a>WSL2 (for Windows)
Linux based operating systems must be used for DYNAMOS. If Windows must be used, the Windows Subsystem for Linux (WSL2) must be used.

Please refer to Windows documentation on how to enable this. Here is one example of a tutorial:

https://learn.microsoft.com/en-us/windows/wsl/install

##  2. <a name='DockerDesktop'></a>Docker Desktop
For local development we utilise Docker Desktop, as upon installation we get Kubernetes for free in an easily manageable interface.

Need to enable kubernetes in order to install some of the following requirements. On docker desktop this can be done by following Settings > Kubernetes > Enable Kubernetes. Check that both Kubernetes client and server:
```bash
kubectl version
```

Using Minikube is not recommended, it should technically work but has its own set of challenges is not supported in these documents.

**NOTE**: In our development cycle we build and upload "finalized" images to docker hub, thus having a docker hub account may be useful if you intend to further develop services in DYNAMOS.

https://www.docker.com/products/docker-desktop/

Extra instructions:
1. Make sure to enable Kubernetes and provide sufficient resources to Docker when installed. We recommend at least 8GB of RAM and 4 CPU cores.
2. **For Windows users**: Make sure to enable WSL integration in the settings menu. If the option is not there, you may need to edit a script within windows to enable the integration. If you are encountering issues with enabling WSL on Docker, use this blog as a reference: https://docs.docker.com/desktop/wsl/

##  3. <a name='Homebrewoptional'></a>Homebrew (optional)
A package manager that may ease the setup process. This is automatically installed on most MacOS systems, however it requires some setup for Linux.

Homebrew for Linux:

https://docs.brew.sh/Homebrew-on-Linux


Homebrew is not available for Windows, however it is available for WSL.

# Installing
Now that we have our environment setup, we can start installing the required software to deploy DYNAMOS.

##  4. <a name='kubectl'></a>kubectl
https://kubernetes.io/docs/tasks/tools/

Kubernetes CLI tool.

Homebrew:
```bash
brew install kubectl
```

Snap:
```bash
sudo snap install kubectl --classic
```

##  5. <a name='HelmCLI'></a>Helm CLI
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

##  6. <a name='Linkerd'></a>Linkerd
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
linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -

linkerd check

# Install Jaeger onto the cluster for observability
linkerd jaeger install | kubectl apply -f -

# Optionally install for insight dashboard - not currently in use
# linkerd wiz install | kubectl apply -f -
```

##  7. <a name='etcdctloptional'></a>etcdctl (optional)

Required if testing DYNAMOS and you want to delete existing jobs, or check what is in the knowledge base.

```bash
brew install etcd
```

See: https://etcd.io/docs/v3.5/install/

##  8. <a name='k9soptional'></a>k9s (optional)
https://k9scli.io/topics/install/

**Preqeqs**: Homebrew

If you don't want to constantly be using `kubectl` to check the status of your cluster, k9s can be used to visualize all the containers runnings within a namespace and within clusters.

HINT: When running k9s, it will initially load the default namespaces, press 0 to show all namespaces, which your containers are likely going to show up in.

```bash
 brew install derailed/k9s/k9s
```

# System Configuration
Now that we have the required software installed, we can start configuring our system to deploy DYNAMOS to Kubernetes.

##  9. <a name='Installscript'></a>Install script

There is a shell script './configuration/dynamos-configuration.sh' that fully installs all DYNAMOS Helm charts and related configuration. In this section we describe in more detail what this does.

Adjust the DYNAMOS_ROOT path (on the dynamos-configuration.sh file) accordingly.

```
DYNAMOS_ROOT="${HOME}/DYNAMOS"
```

To run the script, execute the following from the root of DYNAMOS

```sh
  ./configuration/dynamos-configuration.sh
```

- Set all requirements for the RabbitMQ password, see section [RabbitMQ password process](#rabbitMQ-password-process)
- Create all namespaces with this RabbitMQ password
- Deploy all services:
  - Ingress
  - Prometheus
  - API Gateway
  - Orchestrations layer
  - Exchange Layer
  - Agents

Understanding this script will help understanding how DYNAMOS is deployed.


##  10. <a name='RabbitMQpasswordprocess'></a>RabbitMQ password process

**This is an explanation of the aforemntioned install script. Do not need to run the commands if the script was successful.**

Every service in DYNAMOS that connects to RabbitMQ requires a user with a password that are configured in RabbitMQ. For now every service has the user 'normal_user' and they share a generic password that we will show how to generate here.

```bash
# NOTE: Below commands are snippets from the full script, for working execution see the full script for all details

# Create a password for a rabbit user
rabbit_pw=$(openssl rand -hex 16)

# Use the RabbitCtl to make a special hash of that password:
rabbiq_mq_hash=$($SUDO docker run --rm rabbitmq:3-management rabbitmqctl hash_password $rabbit_pw)
hashed_pw=$(echo "$rabbiq_mq_hash" | cut -d $'\n' -f2)

# The Rabbit Hashed password needs to be in definitions.json file, that is the configuration for RabbitMQ
sed -i "s|%PASSWORD%|${hashed_pw}|g" ${rabbit_definitions_file}

# Create Kubernetes namespaces and Kubernetes secrets with the generated password
helm upgrade -i -f ${namespace_chart}/values.yaml namespaces ${namespace_chart} --set secret.password=${rabbit_pw}

```

Now:
- The RabbitMQ instance reads 'definitions.json' with the rabbitmqctl hashed PW.
- The actual password is stored as a Kubernetes secret in each namespace, so that services can access it and use it to authenticate with RabbitMQ


###  10.1. <a name='ConfigureRabbitPVC'></a>Configure Rabbit PVC

For a RabbitMQ container to read the definitions.json file the file needs go be uploaded to a Kubernetes PVC. This is done in with the following script, which already done by the main configuration script. Please look through to understand what it does.
```bash
cd configuration
./fill-rabbit-pvc.sh
```

##  11. <a name='Ingress'></a>Ingress

To expose DYNAMOS to a user sending request an 'ingress controller' has been deployed as reverse proxy in the previous step, NGINX in this case. The only service currently exposed from DYNAMOS is the 'API gateway' on port 80.

Since Kubernetes is running locally, the exposed API gateway can be accessed on 'localhost' but does need the correct domain name (api-gateway.api-gateway.svc.cluster.local). For this we edit the hostfile.

###  11.1. <a name='Updatehostfile'></a>Update hostfile

To do this on Linux, use your favourite text editor with root access on the file `/etc/hosts`, like so:

```bash
sudo vim /etc/hosts
```
Now add the following to hosts file:
```bash
127.0.0.1 api-gateway.api-gateway.svc.cluster.local
```

Note that this is required when trying to test DYNAMOS locally using tools such as `curl` or `postman`.


# Bashrc shortcuts

##  12. <a name='AddDYNAMOSenvvarsandhelperfunctionstoshell'></a>Add DYNAMOS env vars and helper functions to shell
To make the deployment process easier, we have prepared a set of environment variables and methods that can be added to your shell rc file. These are usually the `~/.bashrc` or `~/.zshrc` files. Alternatively, the below commands can be added to an additional file, and included in the shell file.

**Don't forget to change the DYNAMOS_ROOT path accordingly.**

```bash
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
(or whatever shell rc is used)

# Example

##  13. <a name='ExampleRequest'></a>Example Request

To make sure we installed everything properly, let's use the AMDeX use case as an example.
Firstly, make sure everything is deployed, which should've been done with the deployment script. Check the status on k9s or using `kubectl` to check the status of all pods method. For shortcuts see [Bashrc shortcuts.](#bashrc-shortcuts)

Let's setup the request.

The URL should be:
```
http://api-gateway.api-gateway.svc.cluster.local:80/api/v1/requestApproval
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

E.g. using curl:
```
curl --location 'http://api-gateway.api-gateway.svc.cluster.local:80/api/v1/requestApproval' \
--header 'Content-Type: application/json' \
--data-raw '{
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
}'
```

# Troubleshooting

##  14. <a name='ServicescrashbecausetheconnectiontoRabbitMQdoesnotwork'></a>Services crash because the connection to RabbitMQ does not work

See [RabbitMQ password process](#rabbitmq-password-process) and/or [Configure Rabbit PVC](#configure-rabbit-pvc)

- Is RabbitMQ running? `kubectl get pods -n core`
- Use K9S or Kubectl to get shell access to RabbitMQ. Check whether there is a `definitions.json` file in `/mnt` with a hashed password
- Check if a Kubernetes secret exists in the namespace of your crashing pod. `kubectl get secret "rabbit" -n <NAMESPACE> -o json | jq -r ".[\"data\"][\"password\"]" | base64 -d`
Note, this password does not match the one in `definition.json`. But you can try hashing this password with rabbitctl and place it in the `defintions.json` to see if this was the issue.
