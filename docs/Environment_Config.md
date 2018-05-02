

# Environment Configuration

Environment Operator utilizes an environment configuration (manifest) file to deploy k8s resources to your namespace.  It defines how to layout environments and which applications (services) to run. This configuration file can either live in your application's repository, or a repository dedicated to just managing this file. At Pearson, we've found our dev teams prefer to manage these files separately from their code repositories. The k8s deployment YAML of Environment operator specifies what repo it is using as well as what the environment file is named. For more information on configuring environment operator and how to specify your environment file, see the [operational guide](./Operatonal_Guide.md). What follows in this document will be information in regards to the configuration of options/features that are available in the manifest file (we will refer to it as environments.bitesize in this document) and how to specify your services for deployment.

<a id="requirements"></a>
#### Requirements:
Environment Operator will deploy containers to Kubernetes minions that have the label shown below. This allows for minions without this label to serve other purposes. For example, if all minions with this label where in a specific networking segment not directly exposed to the internet.

```
role=minion
```
This can be added to your workers/minions with
```
kubectl label nodes <node_name> role=minion
```
----------
<a id="environmentsbitesize"></a>
## environments.bitesize

Here is an example of a complete environments.bitesize manifest:<br>

```
project: docs-dev
environments:
  - name: production
    namespace: docs-dev
    deployment:
      method: rolling-upgrade
    services:
      - name: docs-app-front
        external_url:
          - "kubecon.dev-bite.io"
        backend: app-api
        port: 80
        ssl: "true"
        replicas: 2
        hpa:
          min_replicas: 2
          max_replicas: 5
          target_cpu_utilization_percentage: 75
      - name: docs-app-back
        port: 80
        replicas: 2
        requests:
           cpu: 500m
           memory: 100Mi          
```
We will dive into the configuration, but as a preview, the above configuration when loaded by environment operator will deploy two kubernetes deployments (docs--app-front & docs-app-back), kuberenetes services for those applications, and an ingress for the "docs-app-front" service.

This manifest file consists of:<br>

 * [project name](#projectname)
 * [environments](#environments)
	 * [name](#environmentname)
	 * [deployment method](#deploymentmethod)
	 * [services](#services)<br>


Each environments.bitesize manifest contains building blocks for each environment you intend to deploy/manage. We recommend a consistent naming convention for each namespace (dev, prd, etc) that environment operator will be managing.

<a id="projectname"></a>
**project name**
Naming convention: `<project_name>-<three_letter_env_name>`<br><br>
- Ex. example-dev<br>
- Ex. example-tst<br>
- Ex. example-prd<br>
<br>

<a id="environments"></a>
**environments**
The environment section of the manifest may specify multiple environments to manage.

<a id="environmentname"></a>

 - **name** <br> Every environment starts with a `name`.<br> Along with the name of each environment, we must specify the namespace in which the
   environment deploys to. <br> ```
   - name: production   namespace: docs-dev ```

<a id="deploymentmethod"></a>

 - **deployment method** <br> Currently the only available deployment method is `rolling-update`. A `mode` (optional) can also be specified
   with the deployment method. This is generally used if a manual
   deployment is desired. ``` deployment:   method: rolling-upgrade  
   mode: manual ``` <br>

<a id="services"></a>

 - **services** <br>

   This section of the manifest defines how to provision your service.  Environment operator will create multiple
   Kubernetes resources for that service into your namespace:

    - Kubernetes Service
    - Ingress (Optional)
    - Kubernetes Deployment<br>

   Below are the options that may be specified for each service in the manifest

    - **name** (required): The name of the service that will be created.  This will be the name of the kubernetes service, deployment, and ingress (optional) that will get created by environment operator.
    - **port** (required):  Specifying a port or an array of ports in the manifest provisions a [kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)  into your namespace.  This provides the benefit of DNS resolution of your microservices with the kubernetes ecosystem.
    - **application**: When an application is specified, this corresponds to the docker image name that will be pulled and added as a container within your kubernetes deployment.
    - **version**: This is the version of the docker file that will be pulled.  If a version is specified in your manifest file, the service will be deployed by environment operator immediately.  Services that do not specify a version must be deployed by using the /deploy endpoint of environment-operator.  This provides flexibility for users of environment-operator to decide how/when (automatically versus API request) their deployments are made.
    - **replicas**: This specifies the number of replica pods that will deploy in your kubernetes-deployment. If not specified, this will default to "1"
    - **volumes**: Specifying a volume(s) will create PersistentVolumeClaims within kubernetes that will be mounted into your pod at the path specified. If you wish to manually provision volumes for the PersistentVolumeClaims to bind to, you must set provisioning to "manual" on the volume and create a  corresponding PersistentVolume within kubernetes which has the same name. If you wish to provision volumes dynamically, you may set provisioning to "dynamic" or leave it out as this is the default behaviour. Within Pearson, we EBS volumes provisioned via the aws cloud provider and EFS volumes provisioned via the efs-provisioner. See examples below. Note that the volume would be mounted into each pod that is part of the deployment.
    ```
          services:
          - name: EBS-service
            application: mycontainer
            version: 1
            volumes:
               - name: my-persistent-storage
                 path: /data/mystorage
                 modes: ReadWriteOnce
                 size: 10G
                 type: EBS
          - name: EFS-service
            application: mycontainer
            version: 1
            volumes:
               - name: my-persistent-storage
                 path: /data/mystorage
                 modes: ReadWriteOnce
                 size: 10G
                 type: EFS
          - name: manual-service
            application: mycontainer
            version: 1
            volumes:
               - name: my-persistent-storage (a PV with the same name must exist)
                 path: /data/mystorage
                 modes: ReadWriteOnce
                 size: 10G
                 provisioning: manual
    ```
    - **database_type**: When a database_type is specified (only option supported currently is "mongo") environment-operator will deploy a statefulset into kubernetes for the database. More information on deploying a mongo cluster may be found [here](./Mongo.md)

    - **type**: When a service type is specified, environment operator will create a kubernetes third party resource of the kind specified by this field (CRDs are not currently supported). Further TPR customization (beyond default values) can be specified using the options field for the service. As a working example, within Pearson we use Stackstorm sensors that watch for TPR creation/deletion and trigger Stackstorm workflows which take the options specified as their inputs. 
    ```
        services:
      - name: cb-1
        type: couchbase
        options:
          volume_type: "gp2"
          volume_size: "200"
          instance_type: "t2.large"
          desired_capacity: "1"
          full_backup_sch: "1W:Sun"
          app_id: "100"
          team_id: "dba"
    ```
    - **annotations**: Specifying annotations for your service will add the annotations to the Object Metadata for each pod within your kubernetes deployment. Annotations are an unstructured key/value map that can allow external services to retrieve metadata from your deployment. Pearson is utilizing annotations for scraping of data to Prometheus. Below is an example of how to structure annotations for your service in the manifest:
	```
         annotations:
             - name: random_annotation
	           value: ok_value
    ```

    - **hpa**:   Below is an example of how to specify HPA for your service. In the example below, your deployment would be scaled out to 5 or in to 2 replicas when CPU utilization goes above or below a 75% threshold.  Memory HPA has not been implemented yet within environment-operator. If you are interested in being able to utilize HPA within your kubernetes ecosystem, please review the [requirements](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for HPA in your cluster. In order to specify HPA for your service, you'll need to have Heapster running within your kubernetes ecosystem to gather metrics required for scaling events.
    ```
          services:
          - name: hpaservice
            application: gummybears
            version: 1
            hpa:
               min_replicas: 2
               max_replicas: 5
               target_cpu_utilization_percentage: 75
    ```
    - **limits**:  This is how you specify [limits](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-requests-and-limits-of-pod-and-container) for you service.  If you choose not to specify a limit for your service, the containers that are created will utilize the default limit configuration (1000m CPU/2048MiB Memory) specified by environment operator. This value may be changed within environment operators configuration (pkg>config>config.go). In the example below, the hpaservice pod will be restricted to 500m (.5 CPU core) CPU / 100MiB Memory and will be given Guaranteed QoS.  Since no requests were specified, kubernetees will set the requests equal to the limits. Note: The acceptable unit for CPU in the manifest is "m" and for Memory, "Mi" is supported.  For information on what these units mean, please review the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu).
    ```
         services:
         - name: hpaservice
           application: gummybears
           version: 1
           limits:
              cpu: 500m
              memory: 100Mi
    ```
    - **requests**:  This is how you specify [requests](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-requests-and-limits-of-pod-and-container).  Note: The acceptable unit for CPU in the manifest is "m" and for Memory, "Mi" is supported.  For information on what these units mean, please review the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu).
         - In the example1 service below, the containers that are created for your service will utilize the default limits (1000m CPU/2048MiB Memory) as no limits were specified in the manifest. The service will also have Burstable QoS as the request for 500m (.5 CPU core) CPU / 100MiB Memory is less than the default limit set by environment operator.
         - In the example2 service , the containers that are created will utilize the default limits (1000m CPU/2048MiB Memory) as no limits were specified in the manifest. Additionally, since there were also no requests specified, the requests will be set equal to the default limits, giving the pod Guaranteed QoS of 1000m CPU/2048MiB Memory.
         - Note: For more information on QoS Classes within kubernetes, please review the [documentation](https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/). Environment Operator requires limits to be set on every service, so it supports configuration of services to have either Guaranteed or Burstable QoS.
        ```    
            services:
             - name: example1
               application: gummybears
               version: 1
               requests:
                  cpu: 500m
                  memory: 100Mi
             - name: example2
               application: gummybears
               version: 1
        ```    
    - **external_url**: When one or more external urls are specified, a [kubernetes ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) will be created to allow inbound connectivity to your microservice. Each external_url value will be added as a rule to the ingress object. If this option is omitted, an ingress will not be created.
    - **backend**: By default, the ingress created will direct traffic directly to the service. If you need to change this behaviour, for example to add a proxy layer, you may use this option to do so. It must be set to the value of an existing kubernetes service.  
    - **backend_port**: Used in conjunction with the backend option above. Defaults to the service's "port" value. 
    - **ssl** : Specifying "true" or "false" will result in your Kubernetes Ingress being created with the label "ssl" in its Object Metadata. Pearson utilizes an nginx ingress controller to build out our nginx config for our kubernetes ingresses. When ssl is specified, we ensure that ssl is being utilized when proxing requests to that service. More information on our open sourced nginx controller may be found [here](https://github.com/pearsontechnology/bitesize-controllers).  
    - **env**: This option is not recommended because any change to the environment variables in the manifest file will result in a redeploy of your services.  At pearson, we utilize consul and envconsul for configuring our deployed microservices.  However, this option is available and will allow you to specify environment variables as either variables, k8s secrets or pod fields, that will be available to your pods running in your kubernetes deployment.  In the example below, the "gummybears" container will have access to the VAULT_TOKEN and VAULT_ADDR variables, where contents for one variable is coming from a kubernetes-secret and the other is a specific string.

    ```
          services
          - name: envservice
            application: gummybears
            version: 1
            env:
            - secret: VAULT_TOKEN
              value: vault-glp-dev-read
            - name: VAULT_ADDR
              value: "https://vault.kube-system.svc.cluster.local:8243"
            - name: MY_NODE_NAME
              pod_field: spec.nodeName
    ```
