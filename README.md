# Druid Operator

### Project status: Development 

The project is currently development. 

## Table of Contents

 * [Overview](#overview)

 * [Usage](#usage)    
    * [Install the Operator](#install-the-operator)
    * [Deploy a sample Druid Cluster](#deploy-a-sample-druid-cluster)





### Overview
- This operator deploys a druid cluster, the operator is build using the  [operator-sdk](https://github.com/operator-framework/operator-sdk). 
- Operator shall deploy all the druid process of which MM and Historicals are deployed as Statefulsets whereas all other components are deployed as stateless.

### Install the operator
Register the `Druid` custom resource definition (CRD).

```
$ kubectl create -f deploy/crds/binaryomen.org_druids_crd.yaml
```
Create the operator role and role binding.
Make sure to change the `namespace` acc to convinience.
```
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/all_ns/role.yaml
$ kubectl create -f deploy/service_account.yaml
```
- Run the operator locally
```
NAMESPACE=druid
operator-sdk run --local --namespace $NAMESPACE
```
Verify that the Druid operator is running.
```
adheip@adheip:~/data/operator/druid-operator$ operator-sdk run --local
INFO[0000] Running the operator locally in namespace default. 
{"level":"info","ts":1587823514.0882132,"logger":"cmd","msg":"Operator Version: 0.0.1"}
{"level":"info","ts":1587823514.0882306,"logger":"cmd","msg":"Go Version: go1.14"}
{"level":"info","ts":1587823514.0882354,"logger":"cmd","msg":"Go OS/Arch: linux/amd64"}
{"level":"info","ts":1587823514.0882394,"logger":"cmd","msg":"Version of operator-sdk: v0.16.0"}
{"level":"info","ts":1587823514.0891767,"logger":"leader","msg":"Trying to become the leader."}
{"level":"info","ts":1587823514.0891888,"logger":"leader","msg":"Skipping leader election; not running in a cluster."}
{"level":"info","ts":1587823516.3060777,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":"0.0.0.0:8383"}
{"level":"info","ts":1587823516.3064532,"logger":"cmd","msg":"Registering Components."}
{"level":"info","ts":1587823516.306707,"logger":"cmd","msg":"Skipping CR metrics server creation; not running in a cluster."}
{"level":"info","ts":1587823516.306726,"logger":"cmd","msg":"Starting the Cmd."}
{"level":"info","ts":1587823516.307062,"logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
{"level":"info","ts":1587823516.3072984,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"druid-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1587823516.607953,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"druid-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1587823516.9084291,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"druid-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1587823517.2089837,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"druid-controller"}
{"level":"info","ts":1587823517.2090464,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"druid-controller","worker count":1}
{"level":"info","ts":1587823522.0642064,"logger":"controller_druid","msg":"Reconciling DruidCluster","Request.Namespace":"default","Request.Name":"druid"}
```

### Deploy a sample Druid cluster
```
$ kubectl create -f deploy/crds/cr.yaml
```

Verify that the cluster instances and its components are running.

```
$ kubectl get druid
NAME    AGE
druid   22m
```

```
adheip@adheip:~/data/operator/druid-operator/deploy/crds$ kubectl  get pods
NAME                                 READY   STATUS    RESTARTS   AGE
druid-broker-78d95cb58d-pwcj5        1/1     Running   0          38s
druid-coordinator-5b8799ccdc-wwgf6   1/1     Running   0          36s
druid-historical-0                   1/1     Running   0          43s
druid-middlemanager-0                1/1     Running   0          39s
druid-overlord-74fc4b4f69-dlgjf      1/1     Running   0          40s
druid-router-64499d6498-r4sgr        1/1     Running   0          35s

```
