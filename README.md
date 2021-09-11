# kubenx
kubenx is command line tool combining `eks cli` and `kubectl`
<br><br>

## Installation
- Please Install `go` first
    - https://golang.org/doc/install
```bash
# macos 
$ brew tap GwonsooLee/gslee
$ brew install kubenx
$ kubenx version
Current Version is v1.0.0

# Linux
$ curl -LO https://feeldayone-public.s3.ap-northeast-2.amazonaws.com/release/linux/latest/kubenx.tar.gz
$ gzip -d kubenx.tar.gz
$ tar -xzvf kubenx.tar
$ mv kubenx /usr/local/bin
$ kubenx version
Current Version is v1.0.0
```
<br>

* You can setup auto completion for kubenx command
```bash
# For Bash shell
echo "source <(kubenx completion bash)" >> ~/.bashrc

# For zsh shell
echo "source <(kubenx completion zsh)" >> ~/.zshrc
```

## Multi-Account Context Change
* If you use multiple eks cluster in multiple AWS Account, then you need to assume role to use kubenx.
* Before using, it is `required` to set up a configuration in `$HOME/.kubenx/config`.
    * `session_name` : Role name you want to use for assume. 
    * `assume` : key-value mapping `<Account Alias> : < Assume Role ARN >` 
    * `eks-assume-mapping` : key-value mapping `<Context> : <Account Alias>`
* If you use `kubenx context`, then it will automatically copy assume credentials to clipboard according to the configuration.
    * You should paste to shell by `Ctrl + v`
```bash
{
  "session_name": "Role name you want assume from",
  "assume": {
    "dev" : "arn:aws:iam::22222:role/role-name",
    "stg" : "arn:aws:iam::33333:role/role-name",
    "prod" : "arn:aws:iam::11111:role/role-name",
    "security" : "arn:aws:iam::44444:role/role-name"
  },
  "eks-assume-mapping": {
    "eks-prod-apnortheast2": "prod",
    "eks-dev-apnortheast2": "dev",
    "eks-stg-apnortheast2": "stg",
    "eks-security-apnortheast2": "prod",
  }
}
```
<br> 

## Change Context or Namespace
```bash
$ kubenx context 
Current Context: minikube
? Choose Context:  [Use arrows to move, type to filter]
> minikube
  docker-desktop
  eks-benx-workshop

? Choose Context: docker-desktop
Context is changed to "docker-desktop"
``` 
<br>

```bash
$ kubenx namespace
[ docker-desktop ] Current Namespace: argocd
? Choose Context:  [Use arrows to move, type to filter]
> default
  kube-node-lease
  kube-public
  kube-system
  monitoring

Namespace is changed to "default"

## You can specify the namespace after command
$ kubenx ns kube-system
Namespace is changed to "kube-system"
``` 
<br>

## Only Kubenx can do
### 1. Inspect Node information
* If you want to get detail information about node, you can use this command
* By default, node filters applied is labels with `app`, `env`. If you have these two labels on node, you could easily find the node when running inspect command.
* You could get taint and pod information in the node you choose
* It will only search resources in `all namespaces`. If you want to search in the specific namespace, please use `-n <namespace>` option.
```based
$ kubenx inspect node
? Choose a node:  [Use arrows to move, type to filter]
> docker-desktop (app=nginx, env=test)

? Choose a node: docker-desktop
========Taint INFO=======
purpose=common:NoSchedule

========POD INFO=======
  NAME                               READY  STATUS   HOSTNAME  POD IP      HOST IP       NODE            AGE
  nginx-deployment-56f8998dbc-5jvhr  1/1    Running            10.1.0.171  192.168.65.3  docker-desktop  31m
  nginx-deployment-56f8998dbc-p8xnw  1/1    Running            10.1.0.172  192.168.65.3  docker-desktop  31m
  nginx-deployment-56f8998dbc-pz4b2  1/1    Running            10.1.0.170  192.168.65.3  docker-desktop  32m 
```

### 2. Search Resource by Label
* You can search node and pod resource by label
* You should input `key` and `value` through shell and kubenx will search all nodes and pods with that label
```based
? Key: app
? Value: nginx
Search Selector : app=nginx

========Node INFO=======
No node exists in the namespace
========POD INFO=======
  NAME                               READY  STATUS   HOSTNAME  POD IP      HOST IP       NODE            AGE
  nginx-deployment-56f8998dbc-5jvhr  1/1    Running            10.1.0.171  192.168.65.3  docker-desktop  33m
  nginx-deployment-56f8998dbc-p8xnw  1/1    Running            10.1.0.172  192.168.65.3  docker-desktop  33m
  nginx-deployment-56f8998dbc-pz4b2  1/1    Running            10.1.0.170  192.168.65.3  docker-desktop  34m
  web-0                              0/0    Pending  web-0                                               34m
```

### 3. Clean kubeconfig easily.
* You can clean configurations in kubeconfig. 
* You can select multiple `context` by clicking `space key`.
* Of course you can search context while checking target cluster to delete.
```bash
$ kubenx config delete
? Pick contexts you want to delete:  [Use arrows to move, space to select, type to filter]
  [ ]  eks-sample-apne2
  [ ]  eks-sample-apnortheast2-v2
> [x]  minikube
  [ ]  eks-test-apnortheast2
  [ ]  eks-test2-apnortheast2
  [ ]  eks-common-k8s-useast2
```

### 4. Update kubeconfig from EKS cluster
* You can update kubeconfig without searching eks cluster
```bash
$ kubenx config update
? Choose a cluster:  [Use arrows to move, type to filter]
> eks-sample-apnortheast2-v1
  eks-sample-apnortheast2-v2

Create new context eks-sample-apnortheast2-v1

## Also you can run with cluster name
$ kubenx config update eks-sample-apnortheast2-v1
```
<br>



## Command For EKS Cluster
* Before you search clusters, please assume credentials of target AWS Account
### 1. Initiating VPC Cluster 
* If you create new cluster with terraform, you need to add tag to VPC and subnets.
* Also you manually have to setup OIDC Provider in IAM.
* You can use `cluster init` command which will do these work automatically 
```bash
$ kubenx cluster init
******* Steps for initialization ********
Step 1. Tag setup for VPC
Step 2. Tag setup for public subnet
Step 3. Tag setup for private subnet
Step 4. Create Open ID Connector

? Choose a cluster: eks-stg-apnortheast2-v1
Step 1. VPC Tag needs to be updated
Step 2. Tags for Public Subnet needs to be updated
Step 3. Tags for Private Subnet is already updated
Step 4. New OIDC Provider is successfully created
```
<br>

### 2. Get EKS Cluster
```bash
$ kubenx get cluster
+----------------------------------------------+-------------------------------------------------------------------------------+
|                     NAME                     |                          nginx_example_cluster                                |
+----------------------------------------------+-------------------------------------------------------------------------------+
| Version                                      | 1.15                                                                          |
| Status                                       | ACTIVE                                                                        |
| Arn                                          | arn:aws:eks:<Region ID>:<Account ID>:cluster/<Cluster Name>                   |
| Endpoint                                     | <EKS ENDPOINT>                                                                |
| Cluster SG                                   | <Security Group of Master>                                                    |
| VPC ID                                       | <VPC Name>(<VPC ID>)                                                          |
| VPC Cidr Block                               | <VPC CIDR Block>                                                              |
| <Private Subent1 Name >(<subnet region>)     | <subnet tag1 related to kubernetes>                                           |
|                                              | <subnet tag2 related to kubernetes>                                           |
|                                              | <subnet tag3 related to kubernetes>                                           |
|                                              |                                                                               |
| <Private Subent2 Name >(<subnet region>)     | <subnet tag1 related to kubernetes>                                           |
|                                              | <subnet tag2 related to kubernetes>                                           |
|                                              | <subnet tag3 related to kubernetes>                                           |
+----------------------------------------------+-------------------------------------------------------------------------------+
```
<br>

### 3. Get Node Group 
- You don't need to pass cluster name as argument. You can choose from the terminal
- If you want to pass cluster name as argument, you can specify it with [ --cluster, -c  < cluster name >  ]
```bash
$ kubenx get nodegroup
### Choose Nodegroup
? Choose a nodegroup:  [Use arrows to move, type to filter]
> eks-nginx-node-group1
  eks-nginx-node-group2
  eks-nginx-node-group3

### You can see the information about Nodegroup
? Choose a nodegroup: eks-nginx-node-group1
  NAME                   STATUS  INSTANCE TYPE  LABELS                  MIN SIZE  DISIRED SIZE  MAX SIZE  AUTOSCALING GROUPDS                       DISK SIZE
  eks-nginx-node-group1  ACTIVE  t3.small       app=nginx,env=dev       1         1             1         eks-d8b88e2f-75c2-03c4-6c99-b54b6ad02312  20

  AUTOSCALING GROUP                         INSTANCE ID          HEALTH STATUS  INSTANCE TYPE  AVAILABILITY ZONE
  eks-d8b88e2f-75c2-03c4-6c99-b54b6ad02312  i-041412f0s19f24b1b  Healthy        t3.small       ap-northeast-2c
```


## Kubectl VS kubenx
### 1. Get Current Pod
Kubectl Command
```bash
$ kubectl get pod
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-56f8998dbc-5jvhr   1/1     Running   0          8m15s
nginx-deployment-56f8998dbc-p8xnw   1/1     Running   0          8m14s
nginx-deployment-56f8998dbc-pz4b2   1/1     Running   0          8m46s
web-0                               0/1     Pending   0          8m40s
```

Kubenx Command
- You can see `Pod IP`, `Host IP`, and `the node it is scheduled`.
```bash
$ kubenx get pod
  NAME                               READY  STATUS   HOSTNAME  POD IP      HOST IP       NODE            AGE
  nginx-deployment-56f8998dbc-5jvhr  1/1    Running            10.1.0.171  192.168.65.3  docker-desktop  8m33s
  nginx-deployment-56f8998dbc-p8xnw  1/1    Running            10.1.0.172  192.168.65.3  docker-desktop  8m32s
  nginx-deployment-56f8998dbc-pz4b2  1/1    Running            10.1.0.170  192.168.65.3  docker-desktop  9m4s
  web-0                              0/0    Pending  web-0                                               8m58s
``` 
<br>

### 2. Get Current Service
Kubectl Command
```bash
$ kubectl get service
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   4d8h
nginx        ClusterIP   None         <none>        80/TCP    8m31s
```

Kubenx Command
- You can find `endpoint list` which the service is routing to.
```bash
$ kubenx get service
  NAME        TYPE       CLUSTER-IP  EXTERNAL-IP  PORT(S)  ENDPOINT(S)                        AGE
  kubernetes  ClusterIP  10.96.0.1   <None>       443/TCP  10.1.0.171,10.1.0.172,10.1.0.170,  4d8h
  nginx       ClusterIP  None        <None>       80/TCP   10.1.0.171,10.1.0.172,10.1.0.170,  8m7s
``` 
<br>

### 3. Get Current Deployment
Kubectl Command
```bash
$ kubectl get deployment
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           14m
```

Kubenx Command
- You can find the `deployment strategy` and the configurations about it.
```bash
$ kubenx get deployment
  NAME              READY  UP-TO-DATE  AVAILABLE  STRATEGY TYPE  MAXUNAVAILABLE  MAXSURGE  CONTAINERS  IMAGE        AGE
  nginx-deployment  3      3           3          RollingUpdate  25%             25%       nginx       nginx:1.9.1  14m
``` 
<br>

### 4. Get Current Ingress
Kubectl Command
```bash
$ kubectl get ingresses.
NAME            HOSTS   ADDRESS   PORTS   AGE
ingress-nginx   *                 80      7m53s
```

Kubenx Command
- All ports and paths are shown.
- You can see `target service` in the list.
```bash
$ kubenx get ingress
  NAME           HOST  ADDRESS  PATH  PORTS  TARGET SERVICE  AGE
  ingress-nginx  *              /*    80     service-nginx   7m48s
``` 
<br>

### 5. Port Forward
Kubectl Command
```bash
$ kubectl get pods 
NAME                     READY   STATUS    RESTARTS   AGE
nginx-5578584966-czg6j   1/1     Running   0          100m

$ kubectl port-forward nginx-5578584966-czg6j 8099:80
Forwarding from 127.0.0.1:8099 -> 80
Forwarding from [::1]:8099 -> 80

```

Kubenx Command
- You are able to choose pod from terminal.
- You can cancel anytime via `<Ctrl> + c`.
- Default pod port is the local port you choose.
```bash
$ kubenx port-forward
? Choose a pod:  [Use arrows to move, type to filter]
> nginx-5578584966-czg6j

? Choose a pod: nginx-5578584966-czg6j
? Local port to use: 8080
? Pod port[ Default: 8080]: 80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
Port forwarding is ready to get traffic. have fun!
``` 
<br>


