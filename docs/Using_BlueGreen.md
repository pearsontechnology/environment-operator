# Using Blue/Green with environment-operator

**NOTE:** Needs `environment-operator` running at least `0.0.28` version.

## Overview
`environment-operator` allows you to use blue/green deployment method on a per-service basis. Minimal deployment block in `environments.bitesize` configuration file for the service should look as following:

```
service: my-service
deployment:
  method: bluegreen 		# enables blue-green deployment
  active: blue 	   		# service set serving live traffic (see below)
```

This definition enables blue/green deployment for the specified service, and also tells environment-operator that live traffic should be serviced by "blue" service set. Using the above definition, environment-operator will create and manage three bitesize services:

	* my-service -- this will point to the live service set
	* my-service-blue -- blue service set
	* my-service-green - green service set

Please note, that in blue/green setup environment-operator will be only managing **inactive** service set properties, and will not apply any changes that can impact live traffic. As an example, if we would want to scale up our deployment and set `replicas: 2` in the configuration above, it would only scale `my-service-green` service set and not the active `my-service-blue` service set. If you want to make changes to the live service, you need to make changes to the inactive service set first (by modifying `environments.bitesize` configuration) and then release your change by switching `active:` definition to point to your newly released service.

## Differences between "newstack" blue/green deployment and environment-operator

There are two major differences in how blue/green deployment functions in newstack and via environment-operator:
	1. In "newstack" setup blue/green deployments operate on a ingress resource level. What it means technically, is that only external traffic reaching the cluster is affected by the separation between `blue` and `green` resources. In environment-operator setup, it is done on a service level, which means that it affects not only external traffic but also internal communication between microservices. If you need services to match service set colour, you need to configure your application to be blue/green aware (see below: Determining service colour from the container).
	2. In "newstack" you could only setup blue/green deployment on the environment level, which meant that all services switched over to the other service set at once. In environment-operator setup this setting is more granular and allows you to set deployments on a service level. If you want to update your services at once in the same fashion as in the newstack configuration, you will have to update `active:` deployment reference in all of your services. 

## Deploying service via script
`environment-operator` has two endpoints that can help you to release changes to the environment. Instead of statically pinning your service version in `environments.bitesize` configuration with service's `version: ` directive, you can trigger deployment via webhook to the environment-operator:

```
curl -H 'Content-Type: application/json' \
	   -H 'Authorization: Bearer ${TOKEN} \
	   -d '{"name":"<service_name>", "application":"<appname>", "version": "<appversion>" } \
		${environment_operator_endpoint}/deploy
```

The params in the request are as follows:
	* `${TOKEN}` -- environment-operator authentication token that was configured during environment-operator creation
	* `<service_name>` -- Name of the service to deploy. Must match `name:` configuration for a service  in `environments.bitesize` file.
	* `<appname>` -- Application (image) name to deploy
	* `<version>` - Version to deploy. Matches docker image tag
	* `${environment_operator_endpoint}` -- HTTP endpoint for environment operator.

After this request completes, it is possible to query deployment progress via `/status` endpoint:

```
curl -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer ${TOKEN}' \
     ${environment_operator_endpoint}/status

{
  "environment": "",
  "namespace": "sample-app",
  "services": [
    {
      "name": "back",
      "replicas": {
        "available": 0,
        "up_to_date": 0,
        "desired": 0
      },
      "status": "green"
    },
    {
      "name": "front",
      "version": "latest",
      "deployed_at": "2019-03-24 14:43:02 +0000 UTC",
      "replicas": {
        "available": 3,
        "up_to_date": 3,
        "desired": 3
      },
      "status": "green"
    },
    {
      "name": "front-blue",
      "version": "latest",
      "deployed_at": "2019-03-24 14:43:02 +0000 UTC",
      "replicas": {
        "available": 3,
        "up_to_date": 3,
        "desired": 3
      },
      "status": "green"
    }
  ]
}
```

In the output above, your service's status will be located in `services` block. Output for a "parent" service (`"front"` in the example) will match inactive service set's output.

HTTP requests above are identical to the calls that are made by environment-operator Jenkins plugin, so for Jenkins integration, it is recommended to use a plugin.

## Determining service colour from the container

There are cases where you want to know which service set current container belongs to. It can be used for resource isolation or to manage different configuration properties (for example, if you would like to connect your blue service to other blue services only instead of whichever deployment is active). For this purpose, there is a single environment variable exposed to your container - `POD_DEPLOYMENT_COLOUR` ,  with the value of either `blue` or `green`, depending on the service set current pod belongs to.


## Custom URLs

By default, environment-operator will generate URLs for you blue and green service sets. They will be generated based on a main, active url, and appending service colour to URL's host part. I.e., if service `test` has a setting `external_url: www.external.url`, blue service set can be accessed directly via `www-blue.external.url` and green service set will be served by `www-green.external.url`.

It is possible to override service set's URLs with custom ones from the `environments.bitesize` file under `deployment` section:

```
service: myservice
external_url: www.external.url
deployment:
  method: bluegreen
  active: blue
  custom_urls:
    blue:
    - www.blueservice.com
    green:
    - www.greenservice.com
    - another.url.greenservice.com
```

In the example above, live traffic will be served via `www.external.url`,  blue service will have one custom url `www.blueservice.com` defined, and green service will be available via 2 URLs - `www.greenservice.com` and `another.url.greenservice.com`. Please note, that custom_urls must provide an array of endpoints even if you have only one URL to define.

URLs are defined as ingress resources within Kubernetes and depend on ingress controller for managing them. Every service (myservice, myservice-blue and myservice-green) will have a single ingress defined with a list of hosts this ingress serves and endpoints pointing to the corresponding resources in Kubernetes cluster.

