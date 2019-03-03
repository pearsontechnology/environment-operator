#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if ! which go > /dev/null; then
	echo "golang needs to be installed"
	exit 1
fi

if ! which docker > /dev/null; then
	echo "docker needs to be installed"
	exit 1
fi

: ${IMAGE:?"Need to set IMAGE, e.g. bitesize-registry.default.svc.cluster.local:5000/bitesize/environment-operator"}
IMAGE_TAG=${IMAGE_TAG:-$(git rev-parse HEAD)}
FULL_IMAGE="${IMAGE}:${IMAGE_TAG}"

BUILD_IMAGE="golang:1.10-alpine"

bin_dir="_output/bin"
mkdir -p ${bin_dir} || true

#CC="/usr/local/bin/gcc-6" GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -v -x \
#	--ldflags '-extldflags "-static"' -o ${bin_dir}/environment-operator ./cmd/operator/main.go
echo "**************************************************************"
echo "***************** Running Unit Tests *************************"
echo "**************************************************************"

 docker run --rm -v "$(pwd)":/go/src/github.com/pearsontechnology/environment-operator \
    -w /go/src/github.com/pearsontechnology/environment-operator \
    -e CGO_ENABLED=1 \
    ${BUILD_IMAGE} \
    /bin/sh -c "apk update && apk add git && go test -v ./..."

echo "**************************************************************"
echo "***************** Building Source ****************************"
echo "**************************************************************"

 docker run --rm -v "$(pwd)":/go/src/github.com/pearsontechnology/environment-operator \
  	-w /go/src/github.com/pearsontechnology/environment-operator \
  	-e CGO_ENABLED=1 \
 	${BUILD_IMAGE} \
    go build -v -o ${bin_dir}/environment-operator ./cmd/operator/main.go
#    /bin/sh -c  "apk update && apk add build-base && go build -v -o ${bin_dir}/environment-operator ./cmd/operator/main.go"

echo "**************************************************************"
echo "***************** Building Docker Image and Pushing **********"
echo "**************************************************************"

echo "== Building docker image ${FULL_IMAGE}"
docker build --tag "${FULL_IMAGE}" -f hack/build/Dockerfile .

echo "== Uploading docker image ${FULL_IMAGE}"
docker push "${FULL_IMAGE}"
