# Charts

Helm uses a packaging format called charts. A chart is a collection of files that describe a related set of Kubernetes resources. A single chart might be used to deploy something simple, like a memcached pod, or something complex, like a full web app stack with HTTP servers, databases, caches, and so on.

Charts are created as files laid out in a particular directory tree, then they can be packaged into versioned archives to be deployed.

# Helm Controller

A simple controller built with the Operator SDK that watches for chart CRD's within a namespace and manages installation, upgrades and deletes using Kubernetes jobs.

**Example environments.bitesize**

```
project: pearsontechnology
environments:
  - name: sample-app-environment
    namespace: sample-app
    deployment:
      method: rolling-upgrade
    services:
      - type: HelmChart
        name: akamai-chart
        version: 5.7
        chart: akamai
        repo: https://admin:xxxxxxx@charts.prsn.io
        target_namespace: kube-system
        values_content: |-
          pipeline: not_a_test
          bitesize_environment: glp1
          region: us-east-2
          bitesize_environment_type: pre
        options:
          api_version: helm.kubedex.com/v1
```