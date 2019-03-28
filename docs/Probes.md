# Using Liveness Probes

The kubelet uses liveness probes to know when to restart a Container. For example, liveness probes could catch a deadlock, where an application is running, but unable to make progress. Restarting a Container in such a state can help to make the application more available despite bugs.

# Using Readiness Probes

The kubelet uses readiness probes to know when a Container is ready to start accepting traffic. A Pod is considered ready when all of its Containers are ready. One use of this signal is to control which Pods are used as backends for Services. When a Pod is not ready, it is removed from Service load balancers.

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
        liveness_probe:
          handler:
            tcp_socket:
              port: 8080
          initial_delay_seconds: 5
          period_seconds: 2
        readiness_probe:
          handler:
            exec:
              command:
                - cat
                - /tmp/healthy
          initial_delay_seconds: 2
          period_seconds: 2
```

Both liveness and readiness probes configurations are exatly same as [here](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/) only difference is we need to use `handler` keyword as the above example before mentioning the method of health check.