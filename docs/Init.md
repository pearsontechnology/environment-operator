# Init Containers

A Pod can have multiple Containers running apps within it, but it can also have one or more Init Containers, which are run before the app Containers are started.

* Init Containers are exactly like regular Containers, except:

* They always run to completion.

Each one must complete successfully before the next one is started.
If an Init Container fails for a Pod, Kubernetes restarts the Pod repeatedly until the Init Container succeeds. However, if the Pod has a restartPolicy of Never, it is not restarted.

**Example environments.bitesize**

This will be the simplest configuration to set up init containers.

```
project: sample-app
environments:
  - name: dev
    namespace: sample-app
    deployment:
      method: rolling-upgrade
    services:
      - name: api
        port: 8080
        env:
          - name: API_PORT
            value: 80
        requests:
           cpu: 100m
        init_containers:
          - application: init_one
            name: init1
            version: 1
            command:
              - ls
          - application: init_two
            name: init2
            version: 1
            command:
              - pwd
```

**Example extended environments.bitesize file**

This is how to add sercrets, env and configMaps for init containers.

```
project: sample-app
environments:
  - name: dev
    namespace: sample-app
    gists:
     - name: "application-v1"
       path: "application-v1.bitesize"
       type: configmap
     - name: "application-v2"
       path: "application-v1.bitesize"
       type: configmap
    deployment:
      method: rolling-upgrade
    services:
      - name: api
        version: 1
        ssl: false
        port: 8080
        application: api
        volumes:
          - name: application-v2
            path: "/etc/config/application-v2.config"
            type: configmap
        limits:
          memory: 2048Mi
        replicas: 1
        init_containers:
          - application: init_api_1
            name: init1
            version: 1
            volumes:
              - name: application-v1
                path: "/etc/config/application-v1.config"
                type: configmap
            command:
              - ls
          - application: init_api_2
            name: init2
            version: 1
            command:
              - pwd
            env: 
              - name: sample1
                value: sample1-value
              - secret: sample2
                value: sample2-value
              - name: sample3
                pod_field: metadata.name
```