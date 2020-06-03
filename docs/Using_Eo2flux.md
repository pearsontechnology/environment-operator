# Using Eo2flux migration tool

## Introduction

This tool was developed to ease the migration process from EO to WeaveFlux+Helm environment

It could be used to generate the counterpart Helm release files (one each for each service defined in the EO manifest file) automatically to be used along with the WeaveFlux setup.

It also currently supports to migrate the Consul values from an exported consul key value JSON dump file and exposed them as Env Variables

## How to use

Clone to tool source repository to your local Go environment

```bash
git clone -b eo2flux git@github.com:pearsontechnology/environment-operator.git
```
Generate the compiled binary tool

```bash
go install environment-operator/cmd/eo2flux/eo2flux.go
```

Run the command `eo2flux` with the mandator command line arguments

```bash
./bin/eo2flux -i <EO manifest file> -o <output directory to write the generated files>
```

Optional features 

Migrate the Consul values and expose them as Env Variables

```bash
./bin/eo2flux -i <EO manifest file> -o <output directory to write the generated files> -c <Consul exported value dump JSON file>
```
Example

```bash
./bin/eo2flux -i glp2-qa.bitesize -c glp2-qa-kv.json -o output
``` 

## Current Operational Assumptions when using to migrate the Consul values

- The Consul values should be exported as one value set per Kubernetes namespace

e.g
```bash
consul kv export glp2-qa/ > /tmp/glp2-qa-kv.json
```
- if the key value is defined at the top level without any prefix folder (e.g. `<namespace>/key`), those key values are exposed in all the services considering that they are common to all the services

- if the service is defined in the key according to this format `<namespace>/<service>/key` , those key values for exposed in the relevant service only
- if the service name used in the above format is different from the service name used in the EO manifest file service name, it is expected to map the correct service name used in the Consul as an EnvVar under each service in the EO manifest file with the key name `service_name`
e.g
```bash
   
          - name: service_name
            value: cms-service
``` 
