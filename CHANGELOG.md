# **Change Log**

This project adheres to [Semantic Versioning](http://semver.org/). Additionally, below are the change categories that may be associated with each release version.

- **Added** for new features.
- **Changed** for changes in existing functionality.
- **Deprecated** for once-stable features removed in upcoming releases.
- **Removed** for deprecated features removed in this release.
- **Fixed** for any bug fixes.
- **Security** for any security changes or fixes for vulnerabilities.

### **[1.4.3] [UNRELEASED]**
 #### Changed
  * Fix bug in diff comparison for blue/green services

### **[1.4.2] [RELEASED]**
 #### Changed
  * Add support to K8 1.16

### **[1.4.1] [RELEASED]**
 #### Changed
  * Volumes are sorted before comparing desired and existing state
  * Environment variables from secrets compare properly without the key specified
  * Volumes from secrets are used in comparing desired vs. existing state
  * Added trace logging output
  * Fixed documentation examples for curl

### **[1.4.0] [RELEASED]**
 #### Added
  * Add Aurora CRD support

### **[1.3.9] 2020-03-25 [RELEASED]**
 #### Changed
  * Upgraded k8s.io/client-go to release-11.0 and k8s.io/api to release-1.14 [MAV-270](https://agile-jira.pearson.com/browse/MAV-270)

### **[1.3.8] 2020-03-03 [RELEASED]**
 #### Added
  * Introduced ServiceEntries and fixed some of the istio releated issues.

### **[1.3.7] 2020-02-14 [RELEASED]**
 #### Added
  * Fix changes for externalsecrets always getting detected, which in turn causes problems with HPA by resetting it every 30s.

### **[1.3.6] 2020-01-28 [RELEASED]**
 #### Added
  * Add support for ExternalSecret CRD

### **[1.3.5] 2020-01-22 [RELEASED]**
 #### Fixed
  * Fixed HPA memory metric issue [MAV-59](https://agile-jira.pearson.com/browse/MAV-59)


### **[1.3.4] 2020-01-09 [RELEASED]**
 #### Added
  * Add TLS support for Istio Gateway [BITE-6732](https://agile-jira.pearson.com/browse/BITE-6732)

### **[1.3.3] 2019-12-03 [RELEASED]**
 #### Added
  * Enable Istio Gateway and VirtualService creation [BITE-6209](https://agile-jira.pearson.com/browse/BITE-6209)

### **[1.3.2] 2019-11-26 [RELEASED]**
 #### Added
  * Added DynamoDB Support [BITE-6578](https://agile-jira.pearson.com/browse/BITE-6578)

### **[1.3.1] 2019-09-26 [RELEASED]**
 #### Fixed
  * Fix type case issue with helm controller [BITE-6270](https://agile-jira.pearson.com/browse/BITE-6270)

### **[1.3.0] 2019-09-20 [RELEASED]**
 #### Added
  * Enable dynamic API versions for CRDs [BITE-6010](https://agile-jira.pearson.com/browse/BITE-6010)

### **[1.2.9] 2019-09-05 [RELEASED]**
 #### Fixed
  * replace AWS MSK "Mks" -> "Msks". [BITE-5517](https://agile-jira.pearson.com/browse/BITE-5517)

### **[1.2.8] 2019-08-15 [RELEASED]**
 #### Added
  * Add TLS support for Ingress when `ssl: true` [BITE-5954](https://agile-jira.pearson.com/browse/BITE-5954)

### **[1.2.7] 2019-06-12 [RELEASED]**
 #### Fixed
  * Fixed nil pointer issue on new deployments [BITE-5627](https://agile-jira.pearson.com/browse/BITE-5627)

### **[1.2.6] 2019-06-11 [RELEASED]**
 #### Fixed
  * Fixed changes not getting reflected when HPA is active  [BITE-5627](https://agile-jira.pearson.com/browse/BITE-5627)

### **[1.2.5] 2019-05-29 [RELEASED]**
 #### Fixed
  * Fixed probe "period_seconds" not applied to deployment [BITE-5587](https://agile-jira.pearson.com/browse/BITE-5587)

### **[1.2.4] 2019-05-17 [RELEASED]**
 #### Added
  * Added S3 Support [ODY-120](https://agile-jira.pearson.com/browse/ODY-120)

### **[1.2.3] 2019-05-09 [RELEASED]**
 #### Fixed
  * Fixed a bug with blue/green deployments where external URL change was not reflected correctly

### **[1.2.2] 2019-05-08 [RELEASED]**
 #### Fixed
  * Fixed volume mount issue with init containers [BITE-5404](https://agile-jira.pearson.com/browse/BITE-5404)

### **[1.2.1] 2019-05-03 [RELEASED]**
 #### Fixed
  * Fixed a bug where reaper doesn't remove old HPA config

### **[1.2.0] 2019-04-29 [RELEASED]**
 #### Added
  * Bug fix - Reaper doesn't remove old configs if number of configmaps are 1
  * USE_AUTH environment variable default to true (previous default)

    * `GISTS_USER`
    * `GISTS_TOKEN`
    * `GISTS_PRIVATE_KEY`
  * Update ConfigMap tests to match new defaults
  * Upate ConfigMap docs

### **[1.1.1] 2019-04-26 [RELEASED]**
 #### Added
  * Ability to set LOG_LEVEL to trace

### **[1.1.0] 2019-04-12 [RELEASED]**
 #### Added
  * ConfigMap support for init containers [BITE-5131](https://agile-jira.pearson.com/browse/BITE-5131)

### **[1.0.0] 2019-04-12 [RELEASED]**
 #### Added
  * ConfigMap support [BITE-5048](https://agile-jira.pearson.com/browse/BITE-5048)
  * Updated libraries
  * Updated golang to 1.12
  * Fixed a bug where Git try to pull change when there are no updates available
  * Stable releases

### **[0.0.30] 2019-04-05 [RELEASED]**
 #### Added
  * Option to use github/bitbucket token in place of sshkeys for authentication
  * **NOTE** Default authentication is still based on sshkeys

### **[0.0.29] 2019-04-05 [RELEASED]**
 #### Added
  * Added basic init container support [BITE-5033](https://agile-jira.pearson.com/browse/BITE-5033)

### **[0.0.28] 2019-04-01 [RELEASED]**
 #### Added
  * Added Blue/Green deployment Support [BITE-4946](https://agile-jira.pearson.com/browse/BITE-4946)

### **[0.0.27] 2019-03-27 [RELEASED]**
 #### Added
  * Added liveness and readiness Support [BITE-4993](https://agile-jira.pearson.com/browse/BITE-4993)
  
### **[0.0.26] 2019-03-15 [RELEASED]**
 #### Added
  * Added Sqs Support [ODY-158](https://agile-jira.pearson.com/browse/ODY-158)

### **[0.0.25] 2019-03-05 [RELEASED]**
 #### Added
  * Added HPA based on custom-metrics (BITE-4601)
  * **NOTE** This release breaks backward compatibility for users currently using HPA functionality. If HPA config is not in use and environment-operator fails, please raise an issue.

### **[0.0.24] 2019-02-22 [RELEASED]**
 #### Added
  * Added Sns Support [ODY-140](https://agile-jira.pearson.com/browse/ODY-140)

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
