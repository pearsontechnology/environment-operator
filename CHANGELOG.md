# **Change Log**

This project adheres to [Semantic Versioning](http://semver.org/). Additionally, below are the change categories that may be associated with each release version.

- **Added** for new features.
- **Changed** for changes in existing functionality.
- **Deprecated** for once-stable features removed in upcoming releases.
- **Removed** for deprecated features removed in this release.
- **Fixed** for any bug fixes.
- **Security** for any security changes or fixes for vulnerabilities.

### **[0.0.23] 2019-02-12 [RELEASED]**
 #### Changed
  * bumped k8s.io/client-go to release-8.0 and k8s.io/api to release-1.11 [BITE-4702](https://agile-jira.pearson.com/browse/BITE-4702)
  * bumped k8s.io/apimanchinery to release-1.10

### **[0.0.22] 2019-02-08 [RELEASED]**
 #### Fixed
  * Fix HPA scaleTargetRef API version  [BITE-4676](https://agile-jira.pearson.com/browse/BITE-4676)

### **[0.0.21] 2019-02-06 [RELEASED]**
 #### Added
  * Added support for complex data structures in options [BITE-4641](https://agile-jira.pearson.com/browse/BITE-4641)
  * Added Neptune, Mks, Docdb, Cb support

### **[0.0.20] 2019-01-29 [RELEASED]**
 #### Fixed
  * Fix EO failing when provisioning CRDs [BITE-4545](https://agile-jira.pearson.com/browse/BITE-4545)

### **[0.0.19] 2019-01-17 [RELEASED]**
 #### Fixed
  * Enable continuation of deployments when some deployments are failing due to configs issues [BITE-4428](https://agile-jira.pearson.com/browse/BITE-4428)

### **[0.0.18] 2019-01-16 [RELEASED]**
 #### Changed
  * Upgraded client-go version to v5.0.0 [BITE-4386](https://agile-jira.pearson.com/browse/BITE-4386)

 #### Fixed
  * CRD PUT request failures [BITE-4386](https://agile-jira.pearson.com/browse/BITE-4386)

### **[0.0.17] 2019-01-09 [RELEASED]**
 #### Added
  * Added Postgres Support [BITE-4084](https://agile-jira.pearson.com/browse/BITE-4084)

### **[0.0.16] 2018-09-20 [RELEASED]**
 #### Added
  * Added Zookeeper and Kafka TPRs [BITE-3429](https://agile-jira.pearson.com/browse/BITE-3429)

### **[0.0.15] 2018-09-17 [RELEASED]**

#### Added

* Support for Kubernetes CRDs with backwards compatibilty to k8s 1.7 APIs. [[BITE-3572](https://agile-jira.pearson.com/browse/BITE-3572)]
* Support for mounting secrets as a volume within the container. [[BITE-3581](https://agile-jira.pearson.com/browse/BITE-3581)]

### **[0.0.14] 2018-05-03 [RELEASED]**

#### Added

 * Support different types of dynamic volumes (EBS/EFS) [[BITE-2640](https://agile-jira.pearson.com/browse/BITE-2640)]
 * Added Prometheus metrics endpoint [[BITE-1491](https://agile-jira.pearson.com/browse/BITE-1491)]

### **[0.0.13] 2018-04-26 [RELEASED]**

#### Added

 * Support setting http2 label for ingress objects.  [[BITE-2633](https://agile-jira.pearson.com/browse/BITE-2633)]


 * EO crashes with index out of range when an ingress exists which has no corresponding service defined.

### **[0.0.12] 2018-01-22 [RELEASED]**

#### Changed 

 * Rewrite git pkg using go-git library instead of libgit2 + git2go. 

#### Fixed 

 * EO intermittent panic issue. [[BITE-1941](https://agile-jira.pearson.com/browse/BITE-1941)]  

### **[0.0.11] 2018-01-09 [RELEASED]**

#### Added

 * Support for overriding the default backend for a service's kubernetes ingress.
 * Support for setting pod fields as values for container environment variables. 

#### Fixed

 * Fixed issue with EO trying to update immutable PVC values.
 * Fixed issue with diff being generated when backend_port is not set.

### **[0.0.10] 2017-12-11 [RELEASED]**

#### Added

* Reaper will clean up kubernetes ingress objects that no longer have a corresponding service object external_url configured.

#### Changed

* Defining external_url as a list of values for a service object will now cause a single ingress object with multiple rules to be created instead of multiple ingress objects each with a single rule.

#### Fixed

* EO not taking any action when external_url is defined as a list of values.

### **[0.0.9] 2017-11-28 [RELEASED]**

#### Added

* Added support for configuring multiple external URLs (ingresses) for the same service.
[[BITE-1736](https://agile-jira.pearson.com/browse/BITE-1736)]

#### Changed

*  Persistent volume claims now use dynamic provisioning.
[[BITE-1828](https://agile-jira.pearson.com/browse/BITE-1828)]

### **[0.0.8] - 2017-11-01 [RELEASED]**

#### Added

*  Added mongo support. Environment operator can now stand up a mongodb statefulset if specified in environments.bitesize. [[BITE-1632](https://agile-jira.pearson.com/browse/BITE-1632)]
*  Enabled Guaranteed Quality of Service. Environment operator will now deploy containers with requests=limits when a request is specified within the manifest (environments.bitesize) for a service. [[BITE-1713](https://agile-jira.pearson.com/browse/BITE-1713))]
*  Cleaned up documentation and added a Quick Start Guide. [[BITE-1788](https://agile-jira.pearson.com/browse/BITE-1788))]

### **[0.0.7] - 2017-09-25 [RELEASED]**

#### Fixed

*  Enable unit tests for all environment-operator packages. [[BITE-1472](https://agile-jira.pearson.com/browse/BITE-1472)]
*  Apply/Update services that are only associated with the environment change. [[BITE-1650](https://agile-jira.pearson.com/browse/BITE-1650)]

### **[0.0.6] - 2017-09-13 [RELEASED]**

#### Fixed

*  Ensure k8s resources are only applied if a deployment is made for that Bitesize Service. [[BITE-1634] (https://agile-jira.pearson.com/browse/BITE-1634)]

### **[0.0.5] - 2017-09-06 [RELEASED]**

#### Fixed

* Bug caused by annotations with pods continuously upgrading.

#### Changed

* Service creation logic has changed. Now kubernetes resource will only be created after the deployment fact (i.e. we will not create service, ingress etc. resources for the service that is not yet deployed as a pod)
* (Internals) Pod logs are no longer a part of bitesize environment object.

### **[0.0.4] - 2017-09-01 [RELEASED]**

#### Added

*  Support for kubernetes service annotations. [[BITE-1511](https://agile-jira.pearson.com/browse/BITE-1511)]

### **[0.0.3] - 2017-08-31 [RELEASED]**

#### Added

*  Support for configuring horizontal pod autoscaling. [[BITE-1433](https://agile-jira.pearson.com/browse/BITE-1433)]
*  Added new environment operator endpoint for Pod Status. [[BITE-1484](https://agile-jira.pearson.com/browse/BITE-1484)]
*  Custom Docker registry support added for pod spec. [[BITE-1448](https://agile-jira.pearson.com/browse/BITE-1448)]
*  Environment Operator build/release pipeline now managed by TravisCI. [[BITE-1473](https://agile-jira.pearson.com/browse/BITE-1473)]
*  Add error handling for secrets defined in environment.bitesize files for deployments. [[BITE-1465](https://agile-jira.pearson.com/browse/BITE-1465)]

### **[0.0.2] - 2017-01-17 [RELEASED]**

* Validator command added for validation of environment.bitesize file

### **[0.0.1] - 2017-01-17 [RELEASED]**

* Original release of environment operator.
