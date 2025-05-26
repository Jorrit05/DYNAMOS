#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Install CLI
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge | sh

# Add Linkerd to PATH for this session
export PATH=$HOME/.linkerd2/bin:$PATH

kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml

# Install Linkerd on cluster
linkerd install --crds | kubectl apply -f -
linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -

# Check Linkerd status
linkerd check

# Install Jaeger onto the cluster for observability
linkerd jaeger install | kubectl apply -f -

# Optionally install for insight dashboard - not currently in use
# linkerd wiz install | kubectl apply -f -