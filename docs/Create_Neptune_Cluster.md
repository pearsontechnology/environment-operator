# How AWS Neptune is provisioned in Bitesize

The AWS Neptune cluster can be provisioned in Bitesize as Kubernetes CRD-Custom Resource Object. You need to define a yaml structure as stated below in kubernetes using EO-Environment Operator.

```yaml
apiVersion: prsn.io/v1
kind: Neptune
metadata:
  labels:
    creator: pipeline
    name: susanthab
  name: nep01
  namespace: tpr-dev
spec:
  options:
    ApplyImmediately: True
    EnableCloudwatchLogsExports:
      - "audit"
    DBClusterParameterGroupName: "db-test01-dev-neptune1"
    db_instances:
      - db_name: "db01"
        db_instance_class: "db.r4.2xlarge"
      - db_name: "db02"
        db_instance_class: "db.r4.xlarge"
      - db_name: "db03"
        db_instance_class: "db.r4.large"
      - db_name: "db04"
        db_instance_class: "db.r4.large"
```
The yaml file above creates a Neptune cluster with four (4) db instances. Usually the first db instance (db01) will become writer and all other db instances become replicas. It is entirely possible to create a cluster with just one db instance, which is only the writer. Optionally you can define the replicas based on the use case. 

You can use following parameters as options:

**ApplyImmediately** - By default, its False.<br/>
**EnableIAMDatabaseAuthentication** - By default, its True.<br/>
**BackupRetentionPeriod** - By default, its set to 10 days.<br/>
**Port** - Set to default port, 8182.<br/>
**StorageEncrypted** - Set to True by default.<br/>
**MultiAZ** - Set to False by default. Since this is a High Availability feature, enable explicitly for production env only.<br/>
**EnableCloudwatchLogsExports** - The list of log types that need to be enabled for exporting to CloudWatch Logs.<br/>
**DBClusterParameterGroupName** - The name of the DB cluster parameter group to associate with this DB cluster.<br/>


The db_name is unique within a Neptune cluster. Note: It takes about 10 - 15 min to create / modify clusters before it can be used by the apps.


