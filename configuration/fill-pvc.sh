#!/bin/bash

# Create the temporary pod
kubectl apply -f temp-pod.yaml

# Wait for the pod to be in the 'Running' state
echo "Waiting for temp-pod to be Running..."
kubectl wait --for=condition=Ready pod/temp-pod --timeout=300s

# Copy local files to the PVC
kubectl cp ./k8s_service_files/definitions.json temp-pod:/mnt/
kubectl cp ./k8s_service_files/rabbitmq.conf temp-pod:/mnt/

# Delete the temporary pod
kubectl delete -f temp-pod.yaml
