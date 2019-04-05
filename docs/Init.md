# Init Containers

A Pod can have multiple Containers running apps within it, but it can also have one or more Init Containers, which are run before the app Containers are started.

* Init Containers are exactly like regular Containers, except:

* They always run to completion.

Each one must complete successfully before the next one is started.
If an Init Container fails for a Pod, Kubernetes restarts the Pod repeatedly until the Init Container succeeds. However, if the Pod has a restartPolicy of Never, it is not restarted.

**Example environments.bitesize**

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

Please note that these are the only features currently available, users won't be able to use other features just yet.