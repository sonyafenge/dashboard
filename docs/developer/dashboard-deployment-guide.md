# Dashboard deployment guide

This document describes how to setup your development environment.

1. Preparation

Make sure the following software is installed and added to the $PATH variable:

* Curl 7+
* Git 2.13.2+
* Docker 1.13.1+ ([installation manual](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/))
* Golang 1.15.0+ ([installation manual](https://golang.org/dl/))
  * Dashboard uses `go mod` for go dependency management, so enable it with running `export GO111MODULE=on`.
* Node.js 12 and npm 6 ([installation with nvm](https://github.com/creationix/nvm#usage))
* Gulp.js 4+ ([installation manual](https://github.com/gulpjs/gulp/blob/master/docs/getting-started/1-quick-start.md))

2. Clone the repository.

```bash
git clone https://github.com/Click2Cloud-Centaurus/dashboard.git $HOME/dashboard -b dev-21
# TODO while merging PR change it to "git clone https://github.com/CentaurusInfra/dashboard.git $HOME/dashboard -b centaurus
cd $HOME/dashboard
```

3. Install the dependencies:

```
npm ci
```

If you are running commands with root privileges set `--unsafe-perm flag`:

```
npm ci --unsafe-perm
```

4. Make sure your cluster is running, run the command from arktos directory.
```bash
./cluster/kubectl.sh get nodes -A
```
output of above command
```text
NAME               STATUS   ROLES    AGE   VERSION
centaurus-master   Ready    <none>   30m   v0.9.0
```
5. Create a link for kubeconfig file using following command.

If cluster is created using **arktos-up** script, run following command:
```bash
ln -snf /var/run/kubernetes/admin.kubeconfig $HOME/.kube/config
```
If cluster is created using **kube-up** script, run following command:
```bash
ln -snf $HOME/go/src/k8s.io/arktos/cluster/kubeconfig-proxy $HOME/.kube/config
```

**For scale-out architecture**,
Set config directory path:

```bash
export KUBECONFIG_DIR=$(pwd) # default "/opt/centaurus-configs"
```

To change default kubeconfig path(default is `${HOME}/.kube/config`):
```bash
export DEFAULT_KUBECONFIG=$HOME/go/src/k8s.io/arktos/cluster/kubeconfig-proxy
echo 'centaurus-dashboard:kubeconfig ='$DEFAULT_KUBECONFIG' '>> $HOME/dashboard/.npmrc
```
Note: Default file is `$HOME/.kube/config` or pass the argument `--kubeconfig` with path while running `npm run start` command.

Place all your configs in above config directory or in `/opt/centaurus-configs` and should named as shown below
* TP config file name (for TP-1): `kubeconfig.tp-1`
* TP config file name (for RP-1): `kubeconfig.rp-1`

6. Deploy postgres container
```bash
docker run --name postgresql-container -p <db_port>:5432 -e POSTGRES_PASSWORD=<db_password> -d postgres
```
```bash
export POSTGRES_DB=<postgres_db>
export DB_HOST=<host_IP_address>
export DB_PORT=<db_port>
export POSTGRES_USER=<postgres_username>
export POSTGRES_PASSWORD=<password>
```

7. Update the .npmrc and angular.json file in the dashboard directory for bind address and port.

```bash
cd $HOME/dashboard
sudo sed -i '/bind_address/s/^/#/g' $HOME/dashboard/.npmrc
sudo sed -i 's/8080/9443/g' $HOME/dashboard/angular.json
sudo sed -i '0,/RANDFILE/{s/RANDFILE/\#&/}' /etc/ssl/openssl.cnf
echo 'centaurus-dashboard:bind_address = 0.0.0.0 '>> $HOME/dashboard/.npmrc
```

8. To run dashboard:

```bash
npm run start:https --centaurus-dashboard:kubeconfig=$HOME/.kube/config
```
Leave the terminal running.

## To access the dashboard

Dashboard will be accessible on `https://<hostmachine_ip>:9443`

`<machine_ip>`, where `npm run` command is running.

Default credentials are as follows:

username: `centaurus`

password: `Centaurus@123`


