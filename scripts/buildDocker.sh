#!/usr/bin/env bash
set -euo pipefail

TAG="${TAG-""}"
if [ -z "$TAG" ]; then
    TAG="$(git describe --tags | sed 's#^v##g')"
fi

TARGET=("operator" "controller" "client")
for targetName in "${TARGET[@]}"; do
   docker buildx build -f .docker/Dockerfile . --target ${targetName} -t "docker.knockturnmc.com/marauder/${targetName}:${TAG}"
   if [ "${1-""}" == "--push" ]; then
   	docker push "docker.knockturnmc.com/marauder/${targetName}:${TAG}"
   fi
done
