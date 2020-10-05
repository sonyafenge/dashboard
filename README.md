# Centaurus Dashboard

Centaurus Dashboard is a web-based UI for Centaurus clusters. It is evolved from the Kubernetes Dashboard and aims to develop complete support for Centaurus clusters. The dashboard allows users to manage and interact with resources in the cluster.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes/dashboard/blob/master/LICENSE)

## About

This project is based on [Kubernetes Dashboard v.2.0.0-beta4](https://github.com/kubernetes/dashboard/tree/v2.0.0-beta4).

Centaurus clusters are composed from [Arktos](https://github.com/centaurus-cloud/arktos) and [Mizar](https://github.com/centaurus-cloud/mizar), the Dashboard aims to add support for both projects.

The following features are currently under development:

- Multi-Tenancy
- Multi-Region
- VM Pod
- Mizar Integration

![Dashboard UI workloads page](docs/images/dashboard-ui.png)

## Getting Started

**IMPORTANT:** Read the [Access Control](docs/user/access-control/README.md) guide before performing any further steps. The default Dashboard deployment contains a minimal set of RBAC privileges needed to run.

To deploy Dashboard, execute following command:

```sh
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml
```

To access Dashboard from your local workstation you must create a secure channel to your Kubernetes cluster. Run the following command:

```sh
$ kubectl proxy
```

Now access Dashboard at:

[`http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/`](http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/).

## Create An Authentication Token (RBAC)

To find out how to create sample user and log in follow [Creating sample user](docs/user/access-control/creating-sample-user.md) guide.

**NOTE:**

- Kubeconfig Authentication method does not support external identity providers or certificate-based authentication.
- Dashboard can only be accessed over HTTPS
- [Heapster](https://github.com/kubernetes/heapster/) has to be running in the cluster for the metrics and graphs to be available. Read more about it in [Integrations](docs/user/integrations.md) guide.

## Documentation

Dashboard documentation can be found on [docs](docs/README.md) directory which contains:

- [Common](docs/common/README.md): Entry-level overview
- [User Guide](docs/user/README.md): [Installation](docs/user/installation.md), [Accessing Dashboard](docs/user/accessing-dashboard/README.md) and more for users
- [Developer Guide](docs/developer/README.md): [Getting Started](docs/developer/getting-started.md), [Dependency Management](docs/developer/dependency-management.md) and more for anyone interested in contributing

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [**#sig-ui on Kubernetes Slack**](https://kubernetes.slack.com)
- [**kubernetes-sig-ui mailing list** ](https://groups.google.com/forum/#!forum/kubernetes-sig-ui)
- [**Issue tracker**](https://github.com/kubernetes/dashboard/issues)
- [**SIG info**](https://github.com/kubernetes/community/tree/master/sig-ui)
- [**Roles**](ROLES.md)

### Contribution

Learn how to start contribution on the [Contributing Guidline](CONTRIBUTING.md)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## License

[Apache License 2.0](https://github.com/centaurus-cloud/dashboard/blob/master/LICENSE)

---

_Copyright 2020 [The Kubernetes Dashboard Authors](https://github.com/centaurus-cloud/dashboard/graphs/contributors)_
