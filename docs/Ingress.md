# Using TLS with Ingress

When the `ssl: true` is set in the manifest file, It'll create the Ingress referencing a Kubernetes secret for TLS certificate. This secret should be available in the cluster for the ingress controller to apply the TLS certificate for the given external URL. 

There are few ways to create `kubernetes.io/tls` secrets in a cluster. One method is using the following operator. 
https://github.com/pearsontechnology/kubernetes-external-secrets

The above mentioned operator will sync AWS ASM secrets with Kubernetes secrets. It does this by the use of a CRD (ExternalSecret).
