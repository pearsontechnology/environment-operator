# Using ConfigMaps with environment-operator

**NOTE:** Needs `environment-operator` running at least `1.0.0` version.

## Overview
`environment-operator` allows you to use ConfigMaps with standard and bluegreen deployment methods on a per-service basis. Minimal deployment block in `environments.bitesize` configuration file for the service should look as following:

```yaml
project: my-project
environments:
- name: my-env
  namespace: my-env
  gists:
     - name: application-v1
       path: "test/assets/k8s/application-v1.bitesize"
       type: configmap
  services:
  - name: my-service
    volumes:
      - name: application-v1
        path: "/etc/config/application.config"
        type: configmap
```

This definition imports the asset file reside in the manifest files repository relative to the `.bitesize`. The assets files we called as **gists** which are pure Kubernetes ConfigMap yaml files.
Using above definition, Environment Operator will import the Kubernetes yaml file and creates a ConfigMap in Kubernetes cluster with gist name and within the namespace of the environment.

Please note, that ConfigMap Gists will not automatically create unless you specifiy the volumes block using the name of ConfigMap Gist.


## Generating ConfigMap from multiple application configuration files

Environment Operator can generate Kubernetes ConfigMap using a given list of application configuration files.

```yaml
project: my-project
environments:
- name: my-env
  namespace: my-env
  gists:
     - name: "application-v2"
       files:
          - "k8s/application.properties"
          - "k8s/integrations.properties"
       type: configmap
  services:
  - name: my-service
    volumes:
      - name: application-v2
        path: "/etc/config"
        item:
          - key: "application.properties"
            path: "application.config"
          - key: "integrations.properties"
            path: "integrations.config"
```


this definition will create a config map using the files given in the Gist. Each file will be a key-value pair. In this case `application.properties` and `integrations.properties` keys will be created in ConfigMap with associated file content as values.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: application
data:
  application.properties: "application.properties data"
  integrations.properties: "integrations.properties data"
```

it can be used in the volumes block to map these key-value pair to two configuration files inside the container.

```yaml
volumes:
    - name: application-v2
      path: "/etc/config"
      item:
        - key: "application.properties"
          path: "application.config"
        - key: "integrations.properties"
          path: "integrations.config"
```

this block create two configuration files `application.config` and `integrations.config` inside ``/etc/config`` directory.

## ConfigMap deployment process

```
On fist start or newly added Gist

  Gist                   Generate ConfigMap      Create ConfigMap in K8s
  +--------------+       +--------------+       +--------------+
  |              |       |              |       |              |
  |application-v2+-------+application-v2+-------+application-v2|
  |              |       |              |       |              |
  +--------------+       +--------------+       +--------------+

Deploy service

  Deploy service        Check configmap volumes    Pull latest changes
  +--------------+       +--------------+         +-----------------------+
  |              |       |    Volume    |         |application.properties |
  |  my-service  +-------+application-v2+---------+integrations.properties|
  |              |       |              |         |                       |
  +--------------+       +--------------+         +-----------------------+
                                                            |
                                                            |
   Deploy service   Create/Update ConfigMap in K8s    Generate ConfigMap
  +--------------+       +--------------+          +---------------+
  |              |       |              |          |               |
  |  my-service  +-------+application-v2+----------+ application-v2|
  |              |       |              |          |               |
  +--------------+       +--------------+          +---------------+
```

## Determining respective ConfigMap for based on service colour (Blue-Green)

There are cases where you want to manage different configuration properties (for example, if you would like to use your blue configuration when active color is blue).

For this purpose, there is a single environment variable exposed to your container - `POD_DEPLOYMENT_COLOUR` ,  with the value of either `blue` or `green`, depending on the service set current pod belongs to.

Using this environment variable and using `item:` directive of volumes you can create configuration file for each color inside the container.


```yaml
volumes:
    - name: application-v3
      path: "/etc/config"
      item:
        - key: "application-blue.properties"
          path: "application-blue.config"
        - key: "application-green.properties"
          path: "application-green.config"
```


#### Example blue green application config

```yaml
project: my-project
environments:
- name: my-env
  namespace: my-env
  gists:
     - name: "application-v3"
       files:
          - "k8s/application-blue.properties"
          - "k8s/application-blue.properties"
       type: configmap
  services:
  - name: my-service
    volumes:
      - name: application-v3
        path: "/etc/config"
        item:
         - key: "application-blue.properties"
           path: "application-blue.config"
         - key: "application-blue.properties"
           path: "application-green.config"
```

#### Using an external repository for application config

Before getting started environment operator need to configured with following envionment variables

`GISTS_USER` - Git user if the authentication is token based, Overrides `GIT_USER` environment variable
`GISTS_TOKEN` - Git token if the authentication is token based, Overrides `GIT_TOKEN` environment variable
`GISTS_PRIVATE_KEY` - Git private key secret name if the authentication method is key based, Overrides G`IT_PRIVATE_KEY` environment variable


```yaml
project: my-project
environments:
- name: my-env
  namespace: my-env
  gists_repository:
    remote: "https://github.com/pearsontechnology/sample-app-config.git"
    branch: v1.1.0
  gists:
     - name: "application-v3"
       files:
          - "k8s/application-blue.properties"
          - "k8s/application-blue.properties"
       type: configmap
  services:
  - name: my-service
    volumes:
      - name: application-v3
        path: "/etc/config"
        item:
         - key: "application-blue.properties"
           path: "application-blue.config"
         - key: "application-blue.properties"
           path: "application-green.config"
```