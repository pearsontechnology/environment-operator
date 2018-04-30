# **Change Log**

This project adheres to [Semantic Versioning](http://semver.org/). Additionally, below are the change categories that may be associated with each release version.

- **Added** for new features.
- **Changed** for changes in existing functionality.
- **Deprecated** for once-stable features removed in upcoming releases.
- **Removed** for deprecated features removed in this release.
- **Fixed** for any bug fixes.
- **Security** for any security changes or fixes for vulnerabilities.

### **[0.0.14] **

#### Added

 * Added Prometheus metrics endpoint [[BITE-1491](https://agile-jira.pearson.com/browse/BITE-2633)]

#### Changed 
#### Fixed 

### **[0.0.13] 2018-04-26 [RELEASED]**

#### Added

 * Support setting http2 label for ingress objects.  [[BITE-2633](https://agile-jira.pearson.com/browse/BITE-2633)]

#### Changed 
#### Fixed 

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
