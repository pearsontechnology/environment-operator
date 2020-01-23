# How AWS Dynamodb is provisioned in Bitesize

The AWS Dynamodb cluster can be provisioned in Bitesize as Kubernetes CRD-Custom Resource Object. You need to define a yaml structure as stated below in kubernetes using EO-Environment Operator.

apiVersion: prsn.io/v1
kind: Dynamodb
metadata:
  name: <projectname>
  namespace: <namespace>
  labels:
    type: dynamodb
spec:
  region: "us-east-1"
  tablename: "TableName2"
  pri_key_attributename: "name"
  pri_key_attributetype: "S"
  pitr: "true"
  writecapacityunits: 30
  readcapacityunits: 20
 
  optional_spec:
    sort_key_attributename: "age"
    sort_key_attributetype: "N"
    sort_key_keytype: "RANGE"

# Dynamodb configuring backup

Dynamodb provides two types of Dynamodb Backup Options

Dynamodb point-in-time recovery
Dynamodb On-demand backup
1) Dynamodb point-in-time recovery backup
Point-in-time recovery helps protect your DynamoDB tables from accidental write or delete operations. With point-in-time recovery, you don't have to worry about creating, maintaining, or scheduling on-demand backups. With point-in-time recovery, you can restore that table to any point in time during the last 35 days. DynamoDB maintains incremental backups of your table.

Point-in-time Recovery backup has been enabled through the workflow.  you can enable it by set the  variable "pttr :  "true"  " as below. if it is false , it keeps disable. 



apiVersion: prsn.io/v1
kind: Dynamodb
metadata:
  name: <projectname>
  namespace: <namespace>
  labels:
    type: dynamodb
spec:
  region: "us-east-1"
  tablename: "TableName2"
  pri_key_attributename: "name"
  pri_key_attributetype: "S"
  pitr: "true"
  writecapacityunits: 30
  readcapacityunits: 20
 
  optional_spec:
    sort_key_attributename: "age"
    sort_key_attributetype: "N"
    sort_key_keytype: "RANGE"

