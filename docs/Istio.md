# Istio

Cloud platforms provide a wealth of benefits for the organizations that use them. However, there’s no denying that adopting the cloud can put strains on DevOps teams. Developers must use microservices to architect for portability, meanwhile operators are managing extremely large hybrid and multi-cloud deployments. Istio lets you connect, secure, control, and observe services.

At a high level, Istio helps reduce the complexity of these deployments, and eases the strain on your development teams. It is a completely open source service mesh that layers transparently onto existing distributed applications. It is also a platform, including APIs that let it integrate into any logging platform, or telemetry or policy system. Istio’s diverse feature set lets you successfully, and efficiently, run a distributed microservice architecture, and provides a uniform way to secure, connect, and monitor microservices.

**Example environments.bitesize**

```
project: sample-app
environments:
  - name: sample-app-environment
    namespace: sample-istio
    deployment:
      method: rolling-upgrade
    services:
      - name: istio-admin
        application: istio-admin
        version: v1.0.3
        external_url: sample-istio-dev.glp.pearsondev.tech
        service_mesh: enable // This will enable the Service Mesh
        backend: sample-admin // If not mentioned it will use the service name
        port: 80
        replicas: 1
        limits:
          memory: 1024Mi
          cpu: 1000m
        ssl: true
        httpsOnly: true
```