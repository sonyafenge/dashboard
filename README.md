# Centaurus Dashboard

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/centaurus-cloud/dashboard/blob/master/LICENSE)

Centaurus Dashboard is a web-based UI for Centaurus clusters. It is evolved from the Kubernetes Dashboard and aims to develop complete support for Centaurus clusters. The dashboard allows users to manage and interact with resources in the cluster.

![Dashboard UI workloads page](docs/images/dashboard-ui.png)

## About

This project is based on [Kubernetes Dashboard v.2.0.0-beta4](https://github.com/kubernetes/dashboard/tree/v2.0.0-beta4). It's always recommended to follow the docs from the original project for guidance such as project overview, installation and trouble shooting, you can access them [here](https://github.com/kubernetes/dashboard/blob/master/README.md).

## Project overview

The Dashboard is a portal to connect and manage your centaurus clusters. A Centaurus cluster is composed of [Arktos](https://github.com/centaurus-cloud/arktos) and [Mizar](https://github.com/centaurus-cloud/mizar). Arktos is the cluster management system evolved from Kubernetes, it is the core module of the cluster and Mizar is a cloud networking module that works on top of that. The Dashboard interacts mostly with Arktos, and it's required to set up a working Arktos cluster.

## Features and Roadmap

The following features are currently under development or on the road map:

- Multi-Tenancy (🚧)
- Multi-Region (🛣)
- VM Pod (🛣)
- Mizar Integration (🛣)

## Documentation

- [Creating sample user](docs/user/access-control/creating-sample-user.md): Learn how to create a sample user

### Kubernetes Dashboard Documentation

These section is added for quick reference to the original project.

Dashboard documentation can be found on [docs](https://github.com/kubernetes/dashboard/blob/master/docs/README.md) directory which contains:

- [Common](https://github.com/kubernetes/dashboard/blob/master/docs/common/README.md): Entry-level overview
- [User Guide](https://github.com/kubernetes/dashboard/blob/master/docs/user/README.md): [Installation](https://github.com/kubernetes/dashboard/blob/master/docs/user/installation.md), [Accessing Dashboard](https://github.com/kubernetes/dashboard/blob/master/docs/user/accessing-dashboard/README.md) and more for users
- [Developer Guide](https://github.com/kubernetes/dashboard/blob/master/docs/developer/README.md): [Getting Started](https://github.com/kubernetes/dashboard/blob/master/docs/developer/getting-started.md), [Dependency Management](https://github.com/kubernetes/dashboard/blob/master/docs/developer/dependency-management.md) and more for anyone interested in contributing

## License

[Apache License 2.0](https://github.com/centaurus-cloud/dashboard/blob/master/LICENSE)

---

_Copyright 2020 [The Kubernetes Dashboard Authors](https://github.com/centaurus-cloud/dashboard/graphs/contributors)_
