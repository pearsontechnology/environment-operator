# Blue/Green deployments: technical details
Every bitesize service that is marked as a blue/green deployment, `environment-operator`  split into three services:
	1. "Parent" service -- this will be the original service, tracking original service name. It has a blue/green set as a deployment method and points to one of the child services.
	2. "Child" blue service - this service will have all options inherited from the "parent" service, with the following changes:
		* It's `name` will be suffixed with `-blue` suffix
		* It will have custom external URLs. They will be either generated from "parent" service's external URL automatically, or  will be overriden with completely custom set (see user guide section "Custom URLs" for configuration options)
		* It's own upgrade method is `rolling-upgrade`.
		* It will have additional environment variable injected into Kubernetes pod `POD_DEPLOYMENT_COLOUR` with the value `blue`.
	3. "Child" green service -- similar to "blue" service, with the exception that all properties are marked with `green`.

Additionaly, bitesize "child" service objects have a pointer indicating whether current child service is active or not.

Diagrams below show two service states in blue/green setup from Kubernetes resource perspective:

1. Blue/Green deployment when app has `blue` as active deployment

```

  INGRESS                SERVICE                DEPLOYMENT/STATEFULSET
  +--------------+       +--------------+       +--------------+
  |              |       |              |       |              |
  |   app-blue   +-------+   app-blue   +-------+   app-blue   |
  |              |       |              |       |              |
  +--------------+       +--------------+       +--------------+
                                                       |
                                                       |
  +--------------+       +--------------+              |
  |              |       |              |              |
  |   app        +-------+    app       +--------------+
  |              |       |              |
  +--------------+       +--------------+


  +--------------+       +--------------+       +--------------+
  |              |       |              |       |              |
  |  app-green   +-------+   app-green  +-------+  app-green   |
  |              |       |              |       |              |
  +--------------+       +--------------+       +--------------+
```

2. Blue/Green deployment when app has `green` as active deployment

```
INGRESS                SERVICE                DEPLOYMENT/STATEFULSET
+--------------+       +--------------+       +--------------+
|              |       |              |       |              |
|   app-blue   +-------+   app-blue   +-------+   app-blue   |
|              |       |              |       |              |
+--------------+       +--------------+       +--------------+


+--------------+       +--------------+
|              |       |              |
|   app        +-------+    app       +--------------+
|              |       |              |              |
+--------------+       +--------------+              |
                                                     |
                                                     |
+--------------+       +--------------+       +------+-------+
|              |       |              |       |              |
|  app-green   +-------+   app-green  +-------+  app-green   |
|              |       |              |       |              |
+--------------+       +--------------+       +--------------+
```


When `environment-operator` is deployed initially, it will only configure Ingress and Service  Kubernetes resources, but not Deployment/StatefulSet resources (unless they have static `version` set in configuration file). 

Cleanup process (reaper subsystem in environment-operator) will ignore changes for "active" service, running live traffic.

## environment-operator webhooks
Environment-operator webhooks in blue/green setup operate without any changes, with the exception that "parent" service will always match inactive service. For example, triggering `/deploy` webhook will always trigger deployment for inactive service, and `/status` webhook will always inactive service's status. If direct deploy for a specific service "colour" is required, it is possible to trigger deployment by specifying it's full name (`app-blue` or `app-green` instead of `app`) as a parameter.

