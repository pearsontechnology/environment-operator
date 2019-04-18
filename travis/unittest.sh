#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

export RELEASE="dev"
export branch=${TRAVIS_BRANCH}
echo "export RELEASE=$RELEASE" >> ~/environment.sh
echo "export branch=$branch" >> ~/environment.sh

echo "**************************************************************"
echo "***************** Running Unit Tests *************************"
echo "**************************************************************"

BUILD_IMAGE="golang:1.12-alpine"

docker run --rm -v "$(pwd)":/go/src/github.com/pearsontechnology/environment-operator \
    -w /go/src/github.com/pearsontechnology/environment-operator \
    ${BUILD_IMAGE} \
    /bin/sh -c "apk update && apk add git gcc musl-dev && go test -v ./..."

