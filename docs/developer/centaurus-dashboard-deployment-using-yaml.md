# Setup a Centaurus cluster(using kube-up) and Centaurus dashboard on GCE

This document outlines the steps to deploy arktos cluster on GCE from remote workstation machine. User will need to run following steps on workstation machine (the recommended instance size should be atleast ```16 CPU and 32GB RAM``` and the storage size should be ```150GB``` or more)

### Prerequisites

1. User will need a GCP account, and [gcloud](https://cloud.google.com/sdk/docs/install#deb) configured in your bash profile. Please refer to gcloud configuration documentation or you can use following steps to install and configure gcloud utility.
#### Install gcloud
```bash
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
sudo apt-get install apt-transport-https ca-certificates gnupg -y -q
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
sudo apt-get update -y && sudo apt-get install make google-cloud-sdk -y
gcloud init # Provide credentials
```
2. User will need docker and golang to create binaries and docker images.

#### Install Docker
```bash
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update -y
# Install docker and docker-compose
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose -y
groupadd docker
sudo usermod -a -G docker $USER
sudo service docker restart
sudo init 6 # restart vm
```

#### Install Golang
```bash
wget https://storage.googleapis.com/golang/go1.15.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.4.linux-amd64.tar.gz
sudo echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
sudo echo 'export GOPATH=$HOME/gopath' >> $HOME/.profile
source $HOME/.profile
```

### Deploy Arktos cluster

#### Clone arktos repository
```bash
mkdir -p $HOME/go/src/k8s.io
cd $HOME/go/src/k8s.io
git clone -b poc-2022-01-30 https://github.com/CentaurusInfra/arktos.git
```
#### Build the Arktos release binaries from a bash terminal from your Arktos source root directory
```cgo
cd $HOME/go/src/k8s.io/arktos
make clean
make quick-release
```

#### To deploy the arktos cluster on GCE, run kube-up script as follows:
```bash
unset KUBE_GCE_MASTER_PROJECT KUBE_GCE_NODE_PROJECT KUBE_GCI_VERSION  KUBE_GCE_MASTER_IMAGE  KUBE_GCE_NODE_IMAGE KUBE_CONTAINER_RUNTIME NETWORK_PROVIDER DISABLE_NETWORK_SERVICE_SUPPORT
export SCALEOUT_CLUSTER=true NUM_NODES=3 SCALEOUT_TP_COUNT=2 SCALEOUT_RP_COUNT=2 RUN_PREFIX=test-001
export MASTER_DISK_SIZE=200GB MASTER_ROOT_DISK_SIZE=200GB KUBE_GCE_ZONE=us-central1-b MASTER_SIZE=n2-standard-16 NODE_SIZE=n2-standard-16 NODE_DISK_SIZE=256GB GOPATH=$HOME/go KUBE_GCE_ENABLE_IP_ALIASES=true KUBE_GCE_PRIVATE_CLUSTER=true CREATE_CUSTOM_NETWORK=true KUBE_GCE_INSTANCE_PREFIX=${RUN_PREFIX} KUBE_GCE_NETWORK=${RUN_PREFIX} ENABLE_KCM_LEADER_ELECT=false ENABLE_SCHEDULER_LEADER_ELECT=false ETCD_QUOTA_BACKEND_BYTES=8589934592 SHARE_PARTITIONSERVER=false LOGROTATE_FILES_MAX_COUNT=200 LOGROTATE_MAX_SIZE=200M KUBE_ENABLE_APISERVER_INSECURE_PORT=true KUBE_ENABLE_PROMETHEUS_DEBUG=true KUBE_ENABLE_PPROF_DEBUG=true TEST_CLUSTER_LOG_LEVEL=--v=2 HOLLOW_KUBELET_TEST_LOG_LEVEL=--v=2 GCE_REGION=us-central1-b
# Run kube-up with Mizar as CNI plugin
NETWORK_PROVIDER=mizar ./cluster/kube-up.sh
```

kube-up script displays the admin cluster details upon successful deployment.

### Using the arktos cluster using kubectl

To use arktos cluster, use kubectl utility, e.g:
```bash
./cluster/kubectl.sh get pods --kubeconfig cluster/kubeconfig-proxy
./cluster/kubectl.sh get pods --kubeconfig cluster/kubeconfig.tp-1
./cluster/kubectl.sh get pods --kubeconfig cluster/kubeconfig.rp-1
./cluster/kubectl.sh get pods --kubeconfig cluster/kubeconfig.tp-2
./cluster/kubectl.sh get pods --kubeconfig cluster/kubeconfig.rp-2
```

## Deploy dashboard
```bash
cd $HOME/go/src/k8s.io/arktos
mkdir certs
openssl genrsa -out certs/dashboard.key 2048
openssl rsa -in certs/dashboard.key -out certs/dashboard.key
openssl req -sha256 -new -key certs/dashboard.key -out certs/dashboard.csr -subj "/CN=$(hostname -i | awk '{print $1}')"
openssl x509 -req -sha256 -days 365 -in certs/dashboard.csr -signkey certs/dashboard.key -out certs/dashboard.crt
wget https://raw.githubusercontent.com/Click2Cloud-Centaurus/Documentation/main/deployment_scripts/docker-compose.yml
docker-compose up -d
```

User can access dashboard on **https://IP-address:9443/#/login** with username `centaurus` and password `Centaurus@123`(Please use IP address of workstation machine, where you have performed all the steps)

![](img.png)


## Arktos cluster tear-down

To terminate arktos cluster, run the following:
```bash
sudo ./cluster/kube-down.sh
```
