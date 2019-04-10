# Environment Quick Start Guide

If you are new to using environment-operator, this document will run you through how to get a simple example up and running, which you can then use to customize environment operator per the [Operational Guide](https://github.com/pearsontechnology/environment-operator/blob/dev/docs/Operatonal_Guide.md) and your deployments per the [User Guide](https://github.com/pearsontechnology/environment-operator/blob/dev/docs/User_Guide.md).  

Prerequistes:

- Execution of following steps on a Kubernetes Cluster running 1.5.7+  (Environment Operator is current tested with 1.5.7)
- jq is installed in your cluster as the run-example.sh script below uses it to parse json: ```sudo apt-get install jq```

*********

### Steps

1)  Clone the Environment Operator Github Repo into your kubernetes cluster:

```
git clone https://github.com/pearsontechnology/environment-operator.git
```

2)  Run the helper script to deploy environment operator and deploy two sample-app pods.  When the script finishes the output should show the environemnt-operator, front and back pods running in the sample-app namespace

```
cd environment-operator/example/sample_app
./run_example.sh

# To deploy using github/bitbucker tokens
# Ensure you update operator-deployment file with base64 encoded
# credentials
./run_example.sh token
```

Now that you have a running example of environment operator with an app deployed, your next step should be to explore the  [Operational Guide](https://github.com/pearsontechnology/environment-operator/blob/dev/docs/Operatonal_Guide.md) and [User Guide](https://github.com/pearsontechnology/environment-operator/blob/dev/docs/User_Guide.md) so you may customize the provided example to fit your needs.
