#!/bin/bash
set -eu

ARGS=( "$@" )
if [ "${#ARGS[@]}" -eq 0 ]; then
    echo "No arguments were provided -- require cache directory and Dockerfile/tag pairs"
    exit 1
elif [ "${#ARGS[@]}" -eq 1 ]; then
    echo "Only cache directory argument was provided -- must have at least one Dockerfile/tag pair"
    exit 1
fi

CACHE_DIR=${ARGS[0]}

ARGS=("${ARGS[@]:1}")
if [ "$(( ${#ARGS[@]} %2 ))" -ne 0 ]; then
    echo "There must be an even number of arguments after the cache directory (Dockerfile/tag pairs)"
    exit 1
fi

IMAGE_CACHE="$CACHE_DIR/cached_image.tar.gz"
if [ ! -e "$IMAGE_CACHE" ]; then
    TAGS=()
    i=0
    while [ "$i" -lt "${#ARGS[@]}" ]; do
        DOCKERFILE_DIR=${ARGS[$i]}
        DOCKER_TAG=${ARGS[$(($i+1))]}
        DOCKER_TAG_CLEAN=${DOCKER_TAG//[^a-zA-Z0-9_.]/-}
        TAGS+=("$DOCKER_TAG")

        cd "$DOCKERFILE_DIR"
        docker build -t "$DOCKER_TAG" .
        i=$(($i+2))
    done

    # save cached images
    echo "Saving cached Docker images to $IMAGE_CACHE..."
    docker save  ${TAGS[@]} | gzip > "$IMAGE_CACHE"
fi
