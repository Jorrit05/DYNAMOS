#!/bin/bash

# Create the temporary pod
kubectl apply -f temp-pod.yaml

# Wait for the pod to be in the 'Running' state
echo "Waiting for temp-pod to be Running..."
kubectl wait --for=condition=Ready pod/temp-pod --timeout=300s -n core
kubectl wait --for=condition=Ready pod/temp-pod --timeout=300s -n orchestrator

# Copy local files to the PVC
kubectl cp ./k8s_service_files/definitions.json temp-pod:/mnt/ -n core
kubectl cp ./k8s_service_files/rabbitmq.conf temp-pod:/mnt/ -n core

# Create a tarball of the files
tar -czvf etcd_files.tar.gz -C ./etcd_launch_files/ .

# Copy the tarball to the pod
kubectl cp etcd_files.tar.gz temp-pod-orch:/mnt -n orchestrator

# Untar the files inside the pod (optional, if you want to unpack the files inside the pod)
kubectl exec -n orchestrator temp-pod-orch -- tar -xzvf /mnt/etcd_files.tar.gz -C /mnt

# Delete the temporary pod
kubectl delete -f temp-pod.yaml
