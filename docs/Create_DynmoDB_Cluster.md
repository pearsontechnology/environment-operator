# How DynmoDB cluster is provisioned in Bitesize

The DynmoDB cluster can be provisioned in Bitesize as Kubernetes CRD-Custom Resource Object. You need to define a yaml structure as stated below in kubernetes using EO-Environment Operator.

```yaml
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
```
You can use the following parameters as options:

**region** - Region name.<br/>
**tablename** - Table names must be between 3 and 255 characters long.<br/>
**pri_key_attributename** - This is required.<br/>
**pri_key_attributetype** - This is only support with the String(S),Binary(B) and Number(N).<br/>
**pitr** - Dynamodb point-in-time recovery to enable/ disable.<br/>
**writecapacityunits** - By default 10.<br/>
**readcapacityunits** - By default 10.<br/>
**sort_key_attributename** - This is optional.<br/>
**sort_key_attributetype** - This is only support with the String(S),Binary(B) and Number(N).<br/>
**sort_key_keytype** - This is only support with the 'HASH'|'RANGE'.<br/>
