# BITE-4601 Demo Autoscaling based on custom metrics

This demo showcases Horizontal Pod Autoscaling based on custom metrics

We will use environment-operator to deploy an application to the cluster

environment-operator is running in the sample-app namespace:

```
root@master-i-01d298add324432ca:~/environment-operator [1]# kubectl get all   -n sample-app
NAME                                       READY     STATUS    RESTARTS   AGE
pod/environment-operator-69564d97f-8f2qx   1/1       Running   0          2m

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/environment-operator   ClusterIP   172.31.81.177   <none>        80/TCP    12h

NAME                                   DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/environment-operator   1         1         1            1           12h

NAME                                             DESIRED   CURRENT   READY     AGE
replicaset.apps/environment-operator-69564d97f   1         1         1         12h
```

Bitesize file used in this demo:

```
project: pearsontechnology
environments:
  - name: sample-app-environment
    namespace: sample-app
    deployment:
      method: rolling-upgrade
    services:
      - name: front
        external_url: front.sample-app.domain
        ssl: false
        port: 80
        env:
          - name: APP_PORT
            value: 80
          - name: BACK_END
            value: back.sample-app.svc.cluster.local
        limits:
          cpu: 500m
          memory: 100Mi
      - name: back
        port: 80
        replicas: 1
        hpa:
          min_replicas: 1
          max_replicas: 5
          metric:
            name: cpu
            target_cpu_utilization_percentage: 75
        env:
          - name: APP_PORT
            value: 80
        requests:
           cpu: 100m
        limits:
           cpu: 500m
           memory: 100Mi
      - name: api
        port: 9898
        replicas: 1
        hpa:
          min_replicas: 1
          max_replicas: 10
          metric:
            name: http_requests
            target_average_value: 2
        env:
          - name: APP_PORT
            value: 9898
        annotations:
          - name: prometheus.io/scrape
            value: true
          - name: prometheus.io/path
            value: /metrics
          - name: prometheus.io/port
            value: 9898
        requests:
           cpu: 100m
        limits:
           cpu: 500m
```

We will scale based on the custom metric http_requests – number of http requests per second. To view all the metrics retrieved from prometheus:

```
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "custom.metrics.k8s.io/v1beta1",
  "resources": [
    {
      "name": "pods/nodejs_external_memory_bytes",
      "singularName": "",
      "namespaced": true,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "namespaces/vault_core_handle_request",
      "singularName": "",
      "namespaced": false,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "pods/vault_core_pre_seal_count",
      "singularName": "",
      "namespaced": true,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "pods/vault_route_create_secret__count",
      "singularName": "",
      "namespaced": true,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    },
   -----truncated------
  ]
}   

```

To view the  http_requests metric for our sample-app
```
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/sample-app/pods/*/http_requests" | jq .
{
  "kind": "MetricValueList",
  "apiVersion": "custom.metrics.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/apis/custom.metrics.k8s.io/v1beta1/namespaces/sample-app/pods/%2A/http_requests"
  },
  "items": [
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "sample-app",
        "name": "api-d749d8548-zhfkd",
        "apiVersion": "/v1"
      },
      "metricName": "http_requests",
      "timestamp": "2019-03-21T14:48:41Z",
      "value": "66m"
    }
  ]
}
```

Deploy our app using curl to the environment-operator:

```
kubectl  exec -it  environment-operator-dd5dc4464-p25sn sh  -n sample-app

$ curl -k  -XPOST -H 'Content-Type: application/json' -d '{"application":"sample-app
-podinfo", "name":"api", "version":"v0.0.3"}'  environment-operator.sample-app.svc.clu
ster.local/deploy
{"status":"deploying"}

kubectl get pod -n sample-app
```

Add load to the deployed app:
```
kubectl  exec -it  environment-operator-dd5dc4464-p25sn sh  -n sample-app
 kubectl get svc  api -n sample-app
 NAME      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
 api       ClusterIP   172.31.232.169   <none>        9898/TCP   47m
 $ curl http://api.sample-app.svc.cluster.local:9898
```

Observe HPA scaling stats:
```
 kubectl get  hpa -n sample-app

 kubectl get pod -n sample-app -w

```


