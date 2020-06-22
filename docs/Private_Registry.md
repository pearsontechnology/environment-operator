# Environment operator private registry support 

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
        image: bsomogyi/environment-operator:967f43e1deee2f1c75ab10b6d5b483eee5c58618
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
(index.docker.io/bsomogyi/$app:$version) where app and version come from the environemts.bitesize file.  
 

```
apiVersion: apps/v1
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
          value: bsomogyi
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
        image: bsomogyi/environment-operator:967f43e1deee2f1c75ab10b6d5b483eee5c58618
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



## Bitesize S3 Private Registry

If images are being pulled from the Bitesize S3 Registry (The Bitesize Registry is specific for our Pearson Project) the DOCKER_PULL_SECRETS env variable should be ommited as registry
 secrets are not needed by the pod. We handle access through amazons IAM service. Below is a sample environment-operator yaml file that pulls 
pod images from the Bitesize S3 registry:

```
apiVersion: apps/v1
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
          value: "bitesize-registry.default.svc.cluster.local:5000"
        - name: PROJECT
          value: somogyi-app 
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
        image: bsomogyi/environment-operator:1e25ad1da3de87d4b3e06c99ebd85d618fdb85af
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


## ECR Amazon Elastic Container Registry

For images hosted on ECR, imagePullSecrets are not required. Kubernetes has native support for ECR and kubelet will authenticate with the registry automatically. You must be running in AWS (obviously), have the aws cloud provider enabled and the correct IAM permissions assigned to all you nodes as documented [here](https://kubernetes.io/docs/concepts/containers/images/) (Using AWS EC2 Container Registry section)

You must set EO's DOCKER_REGISTRY environment variable to ACCOUNT.dkr.ecr.REGION.amazonaws.com and ensure that images pushed to this registry follow the naming convention that EO uses to find them. (ACCOUNT.dkr.ecr.REGION.amazonaws.com/project/application/version)
