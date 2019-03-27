# Using Horizontal Pod Autoscaling

Horizontal Pod Autoscaling is a native Kubernetes feature. Horizontal Pod Autoscaling (HPA), allows a developer to dynamically scale the number of pods running in an application depending on CPU utilization, memory or other custom metrics.
Environment operator currently supports scaling the number of pods based on CPU utilization, memory utilization or on custometrics.

## Scaling based on CPU utilization

The following  bitesize file shows an example HPA configuration:


**Example environments.bitesize**

```
project: pidah-app
environments:
  - name: dev
    namespace: pidah-app
    deployment:
      method: rolling-upgrade
    services:
      - name: api
        port: 80
        hpa:
          min_replicas: 1
          max_replicas: 5
	  metric:
	    name: cpu
            target_average_utilization: 80
        env:
          - name: API_PORT
            value: 80
        requests:
           cpu: 100m
```

You specify the minimum and maximum number of pod replicas required by the application, the metric name `cpu` and the threshold – `target_average_utilization` which would trigger a scale up of the replica count. The target_average_utilization is a weighted average across the available number of replicas. Once utilization drops below 80% across the pods, the HPA controller would again dynamically scale down the number of pods in the application.

*NOTE*
It is important to specify the `requests` field above as HPA requires this configuration. Without it, HPA will not work correctly and the pods will not scale dynamically.

## Scaling based on memory utilization

The following is an example config which shows scaling based on memory utilization:


**Example environments.bitesize**

```bash
project: pidah-app
environments:
  - name: dev
    namespace: pidah-app
    deployment:
      method: rolling-upgrade
    services:
      - name: api
        port: 80
        hpa:
          min_replicas: 1
          max_replicas: 5
	  metric:
	    name: memory
            target_average_utilization: 80
        env:
          - name: API_PORT
            value: 80
        requests:
          memory: 128Mi
```

Similar to the cpu utilization example above, you specify the minimum and maximum number of pod replicas required by the application, the metric name `memory` and the threshold – `target_average_utilization` which would trigger a scale up of the replica count.

## Scaling based on custom metrics

Environment Operator currently supports scaling based on custom metrics defined in your application which is exposed to Prometheus. To be able to use this feature, the application developer must first ensure the application is instrumented to expose metrics to Prometheus and then connfigure the desired custom metric in the _environments.bitesize file_.

### Exposing the app metrics to Prometheus
There are prometheus clients avaliable for several languages. The following is a simple example of exposing metrics in a python application. You first install the prometheus python client:

```bash
pip install prometheus_client
```

Then run the following script:
```python
from prometheus_client import start_http_server, Summary
import random
import time

# Create a metric to track time spent and requests made.
REQUEST_TIME = Summary('request_processing_seconds', 'Time spent processing request')

# Decorate function with metric.
@REQUEST_TIME.time()
def process_request(t):
    """A dummy function that takes some time."""
    time.sleep(t)

if __name__ == '__main__':
    # Start up the server to expose the metrics.
    start_http_server(8000)
    # Generate some requests.
    while True:
        process_request(random.random())
```

Now calling the /metrics endpoint for the above script will return the following metrics:
```bash
$ curl localhost:8000/metrics
# HELP python_info Python platform information
# TYPE python_info gauge
python_info{implementation="CPython",major="2",minor="7",patchlevel="15",version="2.7.15"} 1.0
# HELP request_processing_seconds Time spent processing request
# TYPE request_processing_seconds summary
request_processing_seconds_count 45.0
request_processing_seconds_sum 22.108082056045532
# TYPE request_processing_seconds_created gauge
request_processing_seconds_created 1.553735836917498e+09
```

More information on how to expose metrics is available [here](https://prometheus.io/docs/introduction/overview/)

After your application has been instrumented, ensure its packaged and published to a docker registry.

When you are ready for deployment in a cluster you need to add the following annotations section to ensure that prometheus scrapes the app's metrics like this:

```bash
project: pidah-app
environments:
  - name: api
    annotations:
      - name: prometheus.io/scrape
        value: true
      - name: prometheus.io/path
        value: /metrics
      - name: prometheus.io/port
        value: 9898
```
`prometheus.io/scrape` informs prometheus to scrape this application for metrics, `prometheus.io/path` tells prometheus which path to scrape ( this can often be ommited as many applications use the default  /metrics ) and  `prometheus.io/port` informs prometheus which port exposes the metrics.

Next ensure your metrics are available in prometheus. You can do that by going to the Prometheus UI and run a simple query like this:
```
{kubernetes_namespace!=""}		
```
or from the cli you can run something like this:
```bash
curl -k https://prometheus.acme.com/api/v1/query?query='\{kubernetes_namespace!=""\}' | jq

{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "__name__": "gitlab_runner_jobs",
          "app": "default-runner",
          "endpoint": "patch_trace",
          "instance": "ip-10-1-10-236.us-west-2.compute.internal",
          "job": "kubernetes-pods",
          "kubernetes_namespace": "gitlab",
          "kubernetes_pod_name": "default-runner-5d598649bc-4gl5h",
          "pod_template_hash": "1815420567",
          "runner": "m9tKsy2e",
          "status": "202"
        },
        "value": [
          1553518681.029,
          "5"
        ]
      },
      {
        "metric": {
          "__name__": "gitlab_runner_api_request_statuses_total",
          "app": "default-runner",
          "endpoint": "request_job",
          "instance": "ip-10-1-10-236.us-west-2.compute.internal",
          "job": "kubernetes-pods",
          "kubernetes_namespace": "gitlab",
          "kubernetes_pod_name": "default-runner-5d598649bc-4gl5h",
          "pod_template_hash": "1815420567",
          "runner": "m9tKsy2e",
          "status": "201"
        },
        "value": [
          1553518681.029,
          "46"
        ]
      },
      <---- truncated ----->
      ],
    }
  }  
}	
```

You should see your defined metrics in the UI or in the metric `__name__` field in the curl output above.

### Configuring custom metrics

Now that your metrics are in prometheus, you should be able to query your metrics in the cluster as follows:

```bash
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/gitlab/pods/*/gitlab_runner_jobs" | jq

{
  "kind": "MetricValueList",
  "apiVersion": "custom.metrics.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/apis/custom.metrics.k8s.io/v1beta1/namespaces/gitlab/pods/%2A/gitlab_runner_jobs"
  },
  "items": [
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "gitlab",
        "name": "default-runner-5d598649bc-s5dln",
        "apiVersion": "/v1"
      },
      "metricName": "gitlab_runner_jobs",
      "timestamp": "2019-03-27T14:30:36Z",
      "value": "5"
    }
  ]
}
```

The above retrieves the `gitlab_runner_jobs` custom metrics for pods within the `gitlab` namespace.

Once this is confirmed, you can now add the metrics configuration to your _environments.bitesize_ file. The following is a complete example including the annotations and custom metrics definitions:

```bash
project: pidah-app
environments:
  - name: api 
    port: 9898
    replicas: 1
    hpa:
      min_replicas: 1
      max_replicas: 10
      metric:
        name: gitlab_runner_jobs
        target_average_value: 10
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

`target_average_value` is the threshold value that triggers a scale event for the defined metric name `gitlab_runner_jobs`.

## Further Reading

Official documents on HPA is available [here](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)

More information on the custom metrics setup provided by the Prometheus custom metrics Adapter is available [here](https://github.com/DirectXMan12/k8s-prometheus-adapter/blob/master/docs/config-walkthrough.md)

A tutorial on HPA with the prometheus custom metrics adapter is available [here](https://github.com/stefanprodan/k8s-prom-hpa)

A developer's introduction to prometheus is available [here](https://github.com/danielfm/prometheus-for-developers)
