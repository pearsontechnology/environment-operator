---
name: Environment operator operational guide
---

# Environment operator operational guide.

Environment operator runs as a deployment in the namespace it manages. Example deployment (with local keycloak setup) can be found in [example](https://github.com/pearsontechnology/environment-operator/tree/master/example) directory in code repository.

*N.B.* This deployment depends on Keycloak. To use the deployment with static keys, you can ignore `OIDC_ISSUER_URL` and `OIDC_ALLOWED_GROUPS` environment variables in deployment, and setup `AUTH_TOKEN_FILE` environment variable. Recommended way to provide `AUTH_TOKEN_FILE` is via secrets, similar to the method for `GIT_PRIVATE_KEY` (see below).

## Environment variable list

* `GIT_REMOTE_REPOSITORY` - specifies remote repository, where `environments.bitesize` file is located.
* `GIT_PRIVATE_KEY` - git private key, used to authenticate against `GIT_REMOTE_REPOSITORY`. Must allow read-only access.
* `BITESIZE_FILE` - usually `environments.bitesize`, but can be anything, to suit project's needs better (for example, you can have file per environment, or per kubernetes cluster).
* `ENVIRONMENT_NAME` - corresponds to the "name" field in bitesize file. This is the environment that operator manages.
* `DOCKER_REGISTRY` - registry to download application images from.
* `DOCKER_PULL_SECRETS` - A comma delimited list of k8s secret names in your applications k8s namespace that will be used to pull images from your private DOCKER_REGISTRY.
* `PROJECT`  - used for metadata (e.g. tags for managed services).
* `OIDC_ISSUER_URL` - issuer ID for OpenID Connect.
* `OIDC_ALLOWED_GROUPS` - comma separated list of Keycloak provided groups, that can perform HTTP actions against environment-operator.
* `DEBUG` - debug mode.
* `NAMESPACE` - namespace this environment-operator actions on. Usually self-referenced to local namespace.
* `AUTH_TOKEN_FILE` - path to a static auth token file. Usually injected into environment-operator via kubernetes secret.


## Using kubernetes secrets in environment operator

It is recommended that `GIT_PRIVATE_KEY` would be used as a reference to the secret. Create file named key with private key contents (e.g. cp ~/.ssh/id_rsa key) and create secret git-private-key from it:

```
$ kubectl create secret generic git-private-key --from-file=./key
```

Then you can refer to git-private-key in environment-operator deployment

```

        - name: GIT_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: git-private-key
              key: key

```

Similarly, you can create `deploy-auth-token-file` secret (if you are not using Keycloak) and use it in volume mounts:


```
$ kubectl create secret generic deploy-auth-token-file --from-file=./token
```

and use it as a volume:

```
     containers:
       name: environment-operator
       env:
         - name: AUTH_TOKEN_FILE
           value: /etc/auth/token
       ...
       volumeMounts:
         name: "auth-token"
         mountPath: "/etc/auth"
         readOnly: true
     ...
     volumes:
       name: "auth-token"
       secret:
         secretName: deploy-auth-token-file
```


## Creating Ingress


See [environment-operator examples dir](https://github.com/pearsontechnology/environment-operator/tree/dev/example).


# Private registry support

The environment operator allows Docker images to be deployed into a Kubernetes namespace from private registries like
DockerHub as well as Google Container Registry. This document details the process for configuring the environment operator
to use a private registry as well as how to establish Kubernetes secrets to allow the pods to pull from said registry.


## Google Container Registry

Google Container Registry follows the same environment operator steps, but creating the kubernetes secrets for the registry
is a bit different:

1) Go to the Google Developer Console > IAM & Admin > Service Accounts and click "Create service account"

2) Under "service account" select new and name account "glp-write" or a name signifying the program accessing google cloud

 - Give it the roles:  Project>Editor, Project>Viewer, and
   Project>Service Account Actor

 - Select "Furnish new private key" and    select JSON


3) Create the key and store the file on disk (from here on we assume that it was stored under ~/secret.json)

4) Now login to GCR using Docker from command-line:

```$ docker login -u _json_key -p "$(cat ~/secret.json)" https://gcr.io ```

This will generate an entry for "https://gcr.io" in your ~/.docker/config.json file.

5) Copy the config.json so we can remove other access crews so only https://gcr.io access  remains.  Name the new file  "~/docker-config.json” , remove all new lines and leave only the access for GCR. For example:

{"auths": {"https://gcr.io": { "auth": "<key>","email": “<your email used above”}}}

6) Base64 encode this file:
```  base64 -w 0 ~/docker-config.json ```

This will print a long base64 encoded string'

7) Copy the encoded string and paste it into an image pull secret definition (called ~/pullsecret.yaml) shown below:

```
apiVersion: v1
kind: Secret
metadata:
  name: myregistrykey
  namespace: <namespace>
data:
  .dockerconfigjson: <encoded string>
type: kubernetes.io/dockerconfigjson

```

8) Create the secret: ```kubectl create -f ~/pullsecret.yaml```

9) Add the data to the environment operator deployment yaml (example below):


Note: The following example of environment will pull images from $DOCKER_REGISTRY/$PROJECT/$app:$version
(gcr.io/pearson-techops/$app:$version) where app and version come from the environemts.bitesize file.  
Additionally, you will note the PROJECT_ID from the GCR Project is used in the yaml below.  You can determine your id by
using gcloud:

```
gcloud projects list
```

```
kind: Deployment
metadata:
  labels:
    name: environment-operator
  name: environment-operator
  namespace: somogyi-app
spec:
  replicas: 1
  selector:
    matchLabels:
      name: environment-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: environment-operator
      name: environment-operator
    spec:
      containers:
      - name: environment-operator
        env:
        - name: GIT_REMOTE_REPOSITORY
          value: git@github.com:pearsontechnology/somogyi-temp-test-app.git
        - name: GIT_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: git-private-key
              key: key
        - name: DOCKER_REGISTRY
          value: "gcr.io"
        - name: DOCKER_PULL_SECRETS
          value: myregistrykey
        - name: PROJECT
          value: pearson-techops
        - name: ENVIRONMENT_NAME
          value: dev
        - name: BITESIZE_FILE
          value: environments.bitesize
        - name: AUTH_TOKEN_FILE
          value: /auth/token
        - name: DEBUG
          value: "true"
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: pearsontechnology/environment-operator
        imagePullPolicy: Always
        securityContext:
          runAsUser: 1000
        volumeMounts:
        - name: "auth-token"
          mountPath: "/etc/auth"
          readOnly: true
        - name: "git-key"
          mountPath: "/etc/git"
          readOnly: true
        ports:
        - containerPort: 8080
          protocol: TCP
        resources: {}
        terminationMessagePath: /dev/termination-log
      volumes:
      - name: "auth-token"
        secret:
          secretName: "auth-token-file"
      - name: "git-key"
        secret:
```

## DockerHub

If your container images reside on a private DockerHub registry, below are the steps required to deploy environment
operator to use that registry. Note: Secrets generation below will eventually be part of a ST2 action

1) Setup some Env Variables

```
DOCKER_REGISTRY_SERVER=https://index.docker.io/v2/
DOCKER_USER=Type your dockerhub username, same as when you `docker login`
DOCKER_EMAIL=Type your dockerhub email, same as when you `docker login`
DOCKER_PASSWORD=Type your dockerhub pw, same as when you `docker login`
```

2) Create the secret for your DockerHub Creds in the namespace you will deploy environment operator

```
kubectl create secret docker-registry myregistrykey --namespace=<your namespace> \
  --docker-server=$DOCKER_REGISTRY_SERVER \
  --docker-username=$DOCKER_USER \
  --docker-password=$DOCKER_PASSWORD \
  --docker-email=$DOCKER_EMAIL
```

3) Set the following in your environment operator deployment yaml (example yaml provide below):

```
DOCKER_PULL_SECRETS="myregistrykey"
DOCKER_REGISTRY="index.docker.io"
PROJECT="Docker User Name" in your
```

The DOCKER_PULL_SECRETS gets transformed into Pod imagePullSecrets upon deployment of an application.
The variable DOCKER_PULL_SECRETS supports a comma delimited string of secrets in case you need your pod to utilized
multiple different docker accounts when pulling images. For more information on imagePullSecrets documentation may be
found [here](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod)

Note: The following example of environment will pull images from $DOCKER_REGISTRY/$PROJECT/$app:$version
(index.docker.io/pearsontechnology/$app:$version) where app and version come from the environemts.bitesize file.  


```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    name: environment-operator
  name: environment-operator
  namespace: somogyi-app
spec:
  replicas: 1
  selector:
    matchLabels:
      name: environment-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: environment-operator
      name: environment-operator
    spec:
      containers:
      - name: environment-operator
        env:
        - name: GIT_REMOTE_REPOSITORY
          value: git@github.com:pearsontechnology/somogyi-temp-test-app.git
        - name: GIT_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: git-private-key
              key: key
        - name: DOCKER_REGISTRY
          value: "index.docker.io"
        - name: DOCKER_PULL_SECRETS
          value: myregistrykey
        - name: PROJECT
          value: myproject
        - name: ENVIRONMENT_NAME
          value: dev
        - name: BITESIZE_FILE
          value: environments.bitesize
        - name: AUTH_TOKEN_FILE
          value: /auth/token
        - name: DEBUG
          value: "true"
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: pearsontechnology/environment-operator
        imagePullPolicy: Always
        securityContext:
          runAsUser: 1000
        volumeMounts:
        - name: "auth-token"
          mountPath: "/etc/auth"
          readOnly: true
        - name: "git-key"
          mountPath: "/etc/git"
          readOnly: true
        ports:
        - containerPort: 8080
          protocol: TCP
        resources: {}
        terminationMessagePath: /dev/termination-log
      volumes:
      - name: "auth-token"
        secret:
          secretName: "auth-token-file"
      - name: "git-key"
        secret:
```


## Troubleshooting

This section describes common issues and troubleshooting steps when something does not work.

### Environment operator fails to clone git repository

Usual symptom is these lines in the system output log:

```
time="2017-07-27T11:38:34Z" level=error msg="Git clone error: Failed to authenticate SSH session: Waiting for USERAUTH response"
time="2017-07-27T11:38:34Z" level=error msg="Failed to resolve path '/tmp/repository': No such file or directory"
```

Second line indicates that repository clone has failed (could be wrong GIT_REMOTE_REPOSITORY or GIT_PRIVATE_KEY). However, first line indicates that there was user authentication error (most likely, the wrong GIT_PRIVATE_KEY set).

To verify this is the case, you will need to exec into container:

```
kubectl exec -ti $(kubectl get pods | awk '/environment-operator/{print $1}') /bin/bash
```

Then execute the following commands to verify git clone fails:

```
echo $GIT_PRIVATE_KEY > /tmp/key
chmod 0400 /tmp/key
export GIT_SSH_COMMAND="ssh -i /tmp/key"
git clone $GIT_REMOTE_POSITORY
```

This should give you the error environment-operator encounters.
