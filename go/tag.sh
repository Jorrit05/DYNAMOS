set -e
REPO=jorrit05
TAG="0.1.0"

images=("policy_enforcer" "agent" "anonymize" "orchestrator" "sidecar" "query" "algorithm", "api-gateway")

for image in "${images[@]}"; do
    # If the local image name is different from the remote one, use a mapping
    # Otherwise, assume they are the same

    remote_image=$(echo "dynamos-$image" | sed 's/_/-/g')

    docker tag ${image}:latest $REPO/${remote_image}:$TAG
    docker tag ${image}:latest $REPO/${remote_image}:latest
    docker push $REPO/${remote_image}:$TAG
    docker push $REPO/${remote_image}:latest
done

exit 0
