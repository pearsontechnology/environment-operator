#!/bin/bash

source ~/environment.sh

buildDocker(){
  echo "****************************************************************************"
  echo "***************** Building Docker Image ${FULL_IMAGE} **********************"
  echo "****************************************************************************"

  docker build -f ./hack/build/Dockerfile -t $FULL_IMAGE .

  echo "****************************************************************************"
  echo "***************** Pushing Docker Image ${FULL_IMAGE} ***********************"
  echo "****************************************************************************"

  docker push "${REPO}"
}

bin_dir="_output/bin"
mkdir -p ${bin_dir} || true

docker login -u $DOCKER_USER -p $DOCKER_PASS

echo "*******************************************************************************************"
echo "***************** Building Source *********************************************************"
echo "*******************************************************************************************"

docker run --rm -v "$(pwd)":/go/src/github.com/pearsontechnology/environment-operator \
	-w /go/src/github.com/pearsontechnology/environment-operator \
	-e CGO_ENABLED=1 \
  	geribatai/golang:1.8 \
    go build -v -o ${bin_dir}/environment-operator ./cmd/operator/main.go

REPO=pearsontechnology/environment-operator

if [[ $TRAVIS_BRANCH == *"master"* ]]; then

  FULL_IMAGE="${REPO}:${releaseVersion}"
  buildDocker

elif [ $TRAVIS_PULL_REQUEST != "false" ]; then

  FULL_IMAGE="${REPO}:${TRAVIS_PULL_REQUEST_BRANCH}"
  buildDocker

else

  FULL_IMAGE="${REPO}:${TRAVIS_BRANCH}"
  buildDocker

fi
