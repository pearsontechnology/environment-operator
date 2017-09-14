# environment-operator

## Building new docker image

`hack/build/build.sh` script helps building docker image. It requires you to have
`IMAGE` environment variable set (e.g. `IMAGE=https://hub.docker.com/r/pearsontechnology/example/environment-operator`).The `IMAGE` environment variable tells build.sh where to push the image once built.
By default, image will be tagged with the HEAD commit tag. If you want to override
it and release versioned image, set `IMAGE_TAG` environment variable.  Additionally, travisCI will build docker images
for all PRs against the environment-operator repository. Information on travisCI is documented below.
The build process uses https://hub.docker.com/r/pearsontechnology/golang/ as a containerized go environment for builds.

## TravisCI

**PR Build:** TravisCI will test, compile source, and then build docker images for all PRs opened against the environment-operator repository in Github.
For PRs, the environment-operator image will be tagged with the github branch name (ex. pearsontechnology/environment-operator:$branchname) on Dockerhub.

**Dev Build:** When a PR completd and merged to the "dev" branch, travisCI will test, compile source, and build a new
docker image named: pearsontechnology/environment-operator:dev on Dockerhub

**Master Build:** Once dev branch is successfully built by TravisCI, if a new release tag is present in the changelog, travisCI
will merge dev into the master branch and then test, compile source, and build a docker image named: pearsontechnology/environment-operator:$releaseVersion on Docker hub.

## Releasing a New Version of Environment Operator

* Git Clone the Dev Branch of environment-operator:

```
 git clone --branch=dev git@github.com:pearsontechnology/environment-operator.git /tmp/environment-operator
 cd /tmp/environment-operator
```

* Update environment-operator/CHANGELOG.md to specify a new release
  * example: In the changelog, modify the current candidate to a released version with a date:
    * "**[0.0.6]**" ---> "**[0.0.6] - YYYY-MM-DD [RELEASED]**"

* Update environment-operator/version/version.go  to contain the new release version

* Commit the changes to dev:

```
git add environment-operator/CHANGELOG.md
git add environment-operator/version/version.go
git commit -m "Initiating new release"
git push
```

* After pushing to git, travisCI will build the dev branch (detecting the new release version). Upon success, dev will be merged
to master and tagged with the new release version. A new environment-operator image will also be pushed to Dockerhub as
pearsontechnology/environment-operator:$releaseVersion


## Running Tests

Unit tests are stored next to the source files. Unit tests for each package are executed via the hack/build/build.sh
script above or as a part of every TravisCI build.

## Deploying sample environment-operator

First, you will need to provide private git key that could read from the repository
containing `environments.bitesize` file. Create file named `key` with private key
contents (e.g. `cp ~/.ssh/id_rsa key`) and create secret `git-private-key` from
it:

```
$ kubectl create secret generic git-private-key --from-file=./key
```

Modify repository in `example/operator-deployment.yaml`. Deploy operator:

```
$ kubectl apply -f example/operator-deployment.yaml
```
