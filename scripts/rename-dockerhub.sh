#!/bin/bash

# Define your personal and organizational account names
PERSONAL_ACCOUNT="jorrit05"
ORG_ACCOUNT="dynamos1"

# Define a dictionary mapping old repository names to new repository names
declare -A IMAGE_MAP=(
    ["dynamos-aggregate:latest"]="sql-aggregate:latest"
    ["dynamos-query:latest"]="sql-query:latest"
    ["dynamos-algorithm:latest"]="sql-algorithm:latest"
    ["dynamos-federated-learning:latest"]="fl-federated-learning:latest"
    ["dynamos-model-service:latest"]="fl-model-service:latest"
    ["dynamos-evaluate-service:latest"]="fl-evaluate-service:latest"
    ["dynamos-fl-aggregate:latest"]="fl-aggregate:latest"
    ["dynamos-agent:latest"]="agent:latest"
    ["dynamos-anonymize:latest"]="sql-anonymize:latest"
    ["dynamos-orchestrator:latest"]="orchestrator:latest"
    ["dynamos-policy-enforcer:latest"]="policy-enforcer:latest"
    ["dynamos-sidecar:latest"]="sidecar:latest"
    ["dynamos-api-gateway:latest"]="api-gateway:latest"
    ["dynamos-test:latest"]="test:latest"
)

# Loop through each image in the dictionary
for OLD_IMAGE in "${!IMAGE_MAP[@]}"; do
    docker pull ${PERSONAL_ACCOUNT}/${OLD_IMAGE}
    NEW_IMAGE=${IMAGE_MAP[$OLD_IMAGE]}

    # Tag the image with the new organizational account name and new repository name
    docker tag ${PERSONAL_ACCOUNT}/${OLD_IMAGE} ${ORG_ACCOUNT}/${NEW_IMAGE}

    # Push the image to the organizational account with the new repository name
    docker push ${ORG_ACCOUNT}/${NEW_IMAGE}

    # Optionally, remove the old image from the personal account
    # docker rmi ${PERSONAL_ACCOUNT}/${OLD_IMAGE}
done
