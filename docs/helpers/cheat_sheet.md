openssl rand -hex 12 | docker secret create db_root_password -
openssl rand -hex 12 | docker secret create db_dba_password -
openssl rand -hex 12 | docker secret create rabbitmq_user -
(Get hashed pw by logging into rabbit container and "rabbitmqctl hash_password  <PW>", I think there was another way through the api/definitions. But forgot..
Perhaps starting the service, creating the user manually, copying the hash from the api/definitions.. big brain time)

# Kubernetes
watch -n1 "kubectl get pods --all-namespaces | grep -E '^(uva|surf) '"

kubectl describe pod rabbitmq-575f76fff7-v54pr
kubectl logs rabbitmq-575f76fff7-v54pr
kubectl get events
kubectl exec -it <pod_name> -- /bin/bash

kubectl create secret generic rabbit --from-literal=password=K5vKN2bXI25R+1Jd -n orchestrator
kubectl create secret generic rabbit --from-literal=password=K5vKN2bXI25R+1Jd -n uva
kubectl create secret generic rabbit --from-literal=password=K5vKN2bXI25R+1Jd -n vu
kubectl create secret generic rabbit --from-literal=password=K5vKN2bXI25R+1Jd -n surf
kubectl create secret generic rabbit --from-literal=password=K5vKN2bXI25R+1Jd -n api-gateway

kubectl delete secret rabbit -n orchestrator
kubectl delete secret rabbit -n uva
kubectl delete secret rabbit -n vu
kubectl delete secret rabbit -n surf
kubectl delete secret rabbit -n api-gateway

kubectl create secret generic sql --from-literal=db_root_password=$(openssl rand -base64 12) --from-literal=db_dba_password=$(openssl rand -base64 12) -n core

kubectl get secret "rabbit" -n api-gateway -o json | jq -r ".[\"data\"][\"password\"]" | base64 -d

kubectl exec -it $(kubectl get pods -l app=rabbitmq -o jsonpath='{.items[0].metadata.name}') -- /bin/bash
kubectl get services -n core

# SQL
If the database PW doesn't work, and I changed the root password. This is because the environment variable is ignored since the container will use the existing database on my host machine.

# Logs
docker volume create --name=service_logs

# MONGO
db.auth("root", passwordPrompt() )

# GoLang
go mod init
go get github.com/Jorrit05/GoLib@7f4fdc0293d3af27b39f3a7f811322bcd3e6b9dc

# ETCD
etcdctl --endpoints=http://localhost:30005 get / --prefix
etcdctl --endpoints=http://localhost:30005 del /agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl/ --prefix

- Leader:
etcdctl --endpoints=http://etcd1:2379,http://etcd2:2379,http://etcd3:2379 endpoint status --write-out=table

- all container IP:
docker container ls --filter "name=etcd_cluster" --format "{{.ID}}" | xargs -n1 docker container inspect --format "{{.Name}} {{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}"

## Reading from etcdctl
Examples of useful commands (replace endpoints when necessary)
```sh
# Get all keys
etcdctl --endpoints=http://localhost:30005 get --prefix "" --keys-only

# Delete jobs with a prefix, such as:
etcdctl --endpoints=http://localhost:30005 del /agents/jobs/SURF/jorrit.stutterheim --prefix

# Get a specific value based on a key (based on a prefix)
etcdctl --endpoints=http://localhost:30005 get --prefix "/agents/jobs/SURF"
```

# proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative rabbitMQ.proto

# ARgo
https://github.com/argoproj/argo-workflows/releases/tag/v3.4.8

# Prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

sum(container_memory_usage_bytes{namespace="uva"}) by (namespace)
sum(container_cpu_load_average_10s{namespace="uva"}) by (namespace)

# Testing locally:
Send hardcode message from cmd/test
## Python Query
queue
export DATA_STEWARD_NAME="test"
export DESIGNATED_GRPC_PORT="50053"
export SIDECAR_PORT="50051"
export FIRST="1"
export LAST="0"
export JOB_NAME="test"

# aggregate
export DESIGNATED_GRPC_PORT="50054"
export SIDECAR_PORT="50051"
export FIRST="0"
export LAST="1"

# Algorithm
export DESIGNATED_GRPC_PORT="50054"
export SIDECAR_PORT="50051"
export FIRST="0"
export LAST="1"

# LAST PYTHON SERVICE:
export SIDECAR_PORT=50051
export DESIGNATED_GRPC_PORT=50052
export FIRST=1
export LAST=1
export AMQ_PASSWORD="e3febc96e3060970414ac94b9f0fc020"
export AMQ_USER="normal_user"

# Linkerd
https://linkerd.io/2.13/getting-started/

linkerd install --crds | kubectl apply -f -
linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -
linkerd check

linkerd jaeger install | kubectl apply -f -
linkerd viz install --set grafana.url=grafana.grafana:3000 \
  | kubectl apply -f -

kubectl get -n emojivoto deploy -o yaml \
  | linkerd inject - \
  | kubectl apply -f -

linkerd viz install --set grafana.url=grafana.core.svc.cluster.local:3000 \
  | kubectl apply -f -

linkerd jaeger install --set grafana.url=grafana.core.svc.cluster.local:3000 \
  | kubectl apply -f -

linkerd jaeger uninstall | kubectl delete -f -

## Jaeger dashboard
kubectl port-forward -n linkerd-jaeger service/jaeger 16686:16686
linkerd jaeger dashboard

# Ingress
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
<!-- helm install -f "${coreChart}/ingress-values.yaml" nginx oci://ghcr.io/nginxinc/charts/nginx-ingress -n ingress --version 0.18.0 -->
coreChart=/Users/jorrit/Documents/uva/DYNAMOS/charts/core
helm install -f "${coreChart}/ingress-values.yaml" nginx ingress-nginx/ingress-nginx -n ingress
kubectl get svc --namespace ingress nginx

# RabbitMQ
## See RabbitMQ information
Go to http://localhost:30000/ and login with "guest" as username and password. Here you can view queues and other things.


# Infinite job creation (i.e., jobs removed keep being recreated in k8s)
To delete a job that keeps running (e.g., when it did not finish successfully), you can follow these steps:
1. Describe the pod from the job(s), such as in k9s by pressing d on the pod, or use a command, such as:
```sh
# This describes the k8s pod, which you can use to view the labels
# Replace with pod name and namespace
kubectl describe pod jorrit-stutterheim-3539f61fuva1-lcp8 -n uva
# Or alternatively use the following command to get the jobs in a specific namespace:
kubectl get jobs -n uva
# Or all namespaces (-A is the same as --all-namespaces (use kubectl get jobs --help for explanation)):
kubectl get jobs -A
```
2. Delete the job returned from the previous step (for example in the labels you can see "job-name=jorrit-stutterheim-3539f61fuva1"), then the command would be:
```sh
# This deletes the job from k8s, after which is should not be creating the job again
# Replace with pod name and namespace
kubectl delete job jorrit-stutterheim-3539f61fuva1 -n uva
```
