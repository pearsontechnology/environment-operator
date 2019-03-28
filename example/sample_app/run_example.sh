#!/bin/bash

###################################
##Deploy Environment Operator
###################################

kubectl create ns sample-app
kubectl create secret generic git-private-key --from-file=key=./ro-priv --namespace=sample-app
kubectl create secret generic auth-token-file --from-file=token=./token --namespace=sample-app
kubectl create -f operator-rbac.yaml
kubectl create -f operator-deployment.yaml
kubectl create -f operator-ingress.yaml
kubectl create -f operator-svc.yaml

while [ $(kubectl get pods --namespace=sample-app | grep -i environment-operator | awk '{print $3}') != "Running" ]; do
    echo "Waiting for environemnt-operator deployment to enter a running state"
    sleep 5
done

svcip=$(kubectl get svc -n=sample-app environment-operator -o jsonpath='{.spec.clusterIP}')

###################################
##Deploy Backend Sample App
###################################

sleep 5

echo
echo "Deploying Back End Application"
curl -k -s -XPOST -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' -d '{"application":"sample-app-back", "name":"back", "version":"latest"}'  $svcip/deploy

sleep 3

backstatus=$(curl -k -s -XGET -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' $svcip/status | jq -r '.services[] | select(.name=="back") | .status')

while [ "$backstatus" != "green" ]; do
    echo "Waiting for backend deployment to enter running state before deploying front end app"
    sleep 5
    backstatus=$(curl -k -s -XGET -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' $svcip/status | jq -r '.services[] | select(.name=="back") | .status')
done

###################################
##Deploy Frontend Sample App
###################################

echo
echo "Deploying Front End Application"
curl -k -s -XPOST -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' -d '{"application":"sample-app-front", "name":"front", "version":"latest"}'  $svcip/deploy

sleep 3

frontstatus=$(curl -k -s -XGET -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' $svcip/status | jq -r '.services[] | select(.name=="front") | .status')

while [ "$frontstatus" != "green" ]; do
    echo "Waiting for frontend deployment to enter running state before retrieving deploy status from environment operator"
    sleep 5
    frontstatus=$(curl -k -s -XGET -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' $svcip/status | jq -r '.services[] | select(.name=="front") | .status')
done

###################################
## Get Deployment Status from Environment Operator
###################################

echo
echo "Environment Operator Deployment status:"
curl -k -XGET -H "Authentication: Bearer $( cat token )" -H 'Content-Type: application/json' $svcip/status

echo
echo "Deployed pods in sample-app Namespace:"
kubectl get pods --namespace=sample-app
