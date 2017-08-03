# environment-operator


More information on environment operator can be found in https://docs.google.com/document/d/1VYarE5SepyvVFjkXJfnpM3JAgybxkBT53TZzqsOv_pc



## Building new docker image

`hack/build/build.sh` script helps building docker image. It requires you to have
`IMAGE` environment variable set (e.g. `IMAGE=bitesize-registry.default.svc.cluster.local:5000/core/environment-operator`).
By default, image will be tagged with the HEAD commit tag. If you want to override
it and release versioned image, set `IMAGE_TAG` environment variable.

## Running Tests

Unit tests are stored next to the source files. Unit tests for each package are executed via the hack/build/build.sh script above

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

## Releasing new version

* Tag MASTER branch with release version (e.g. v0.0.1)
* Generate `CHANGELOG.md` by running `reportcopter`. Needs [reportcopter](https://github.com/3zcurdia/reportcopter) on local system.
* Build new docker image (see above).
