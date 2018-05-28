#!/bin/bash

set -e

# Do not rebuild/retest image that we already have.
if [ -n "$TRAVIS_TAG" ]; then
  echo "Found release tag"
  docker pull ${DOCKER_USERNAME}/janna-api:${COMMIT}
  docker tag ${DOCKER_USERNAME}/janna-api:${COMMIT} ${DOCKER_USERNAME}/janna-api:${TRAVIS_TAG}
  docker push ${DOCKER_USERNAME}/janna-api:${TRAVIS_TAG}
  exit 0
fi

make

if [ "$TRAVIS_BRANCH" == "master" ]; then
  make push
  make push TAG=latest
fi
