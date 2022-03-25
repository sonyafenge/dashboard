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
git clone https://github.com/Click2Cloud-Centaurus/dashboard.git $HOME/dashboard -b centaurus
# TODO while merging PR change it to "git clone https://github.com/CentaurusInfra/dashboard.git $HOME/dashboard -b centaurus
cd $HOME/dashboard
```

3. Install the dependencies:

```bash
npm ci
```

If you are running commands with root privileges set `--unsafe-perm flag`:

```bash
npm ci --unsafe-perm
```


4. Build docker image for dashboard service

```bash
cd $HOME/dashboard
npm run build
docker image build -t <image_name> dist/amd64/.
docker push <image_name> # You may need to login to docker account
```
