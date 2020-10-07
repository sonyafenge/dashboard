# Create sample user

User access control allows different tenants to access their own "cluster-space" and assigns different authorization to different users. A default login functionality is already implemented in the Dashboard to provide user access control. When develop locally, simply start the dashboard under https mode `npm run start:https` to enable the login view. By default, Centaurus Dashboard uses Bearer Token to authenticate a user. Bearer Tokens can usually be obtained from service accounts, and in order for the cluster to "recognize" a token as a "user", it is required to bind a service account to a ClusterRoleBinding.

### Creating a Service Account

Create a service account with name `admin-user` in your choice of namespace, here we use `default`:

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: default
```

Simply save the config as `service-acount.yaml` and run `kubectl apply -f service-account.yaml`.

### Edit ClusterRoleBinding

In Centaurus cluster, an `admin-role` and an `admin-role-binding` are automatically created for any tenant:

```
kubectl get clusterrole,clusterrolebinding --all-namespaces

NAME                                               AGE
clusterrole.rbac.authorization.k8s.io/admin-role   2m10s

NAME                                                              AGE
clusterrolebinding.rbac.authorization.k8s.io/admin-role-binding   2m9s
```

We will need to add our `ServiceAccount` to our admin `CusterRoleBinding`, the easiest way is to run

```
kubectl edit clusterrolebinding.rbac.authorization.k8s.io/admin-role-binding
```

and add

```
- kind: ServiceAccount
  name: admin-user
  namespace: default
```

Your complete admin `ClusterRoleBinding` will look similar to:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: "2020-10-02T19:23:32Z"
  hashKey: 4553495368106064663
  name: admin-role-binding
  resourceVersion: "839888708420763649"
  selfLink: /apis/rbac.authorization.k8s.io/v1/clusterrolebindings/admin-role-binding
  tenant: system
  uid: f84ec149-50bc-41db-98fa-70b46abe8b75
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin-role
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: admin
- kind: ServiceAccount
  name: admin-user
  namespace: default
```

### Getting a Bearer Token

When we create a `ServiceAccount`, a bearer token is automatically created as a secret, all we do now is retrieving the token. Execute the following command:

```
kubectl -n default describe secret $(kubectl -n default get secret | grep admin-user | awk '{print $1}')
```

It should print something like:

```
Name:         admin-user-token-f5sc2
Namespace:    default
Tenant:       system
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: admin-user
              kubernetes.io/service-account.uid: db7cd748-f9af-431f-8439-c3cba0a67d59

Type:  kubernetes.io/service-account-token

Data
====
ca.crt:     1310 bytes
namespace:  7 bytes
token:      eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImFkbWluLXVzZXItdG9rZW4tZjVzYzIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiYWRtaW4tdXNlciIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImRiN2NkNzQ4LWY5YWYtNDMxZi04NDM5LWMzY2JhMGE2N2Q1OSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvdGVuYW50Ijoic3lzdGVtIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlZmF1bHQ6YWRtaW4tdXNlciJ9.XsfSh300tfGx9BBnOMQxhR93nybwdZxt3wfJJ-k27Gg4D55WT8O7Ymbv7al8mTi6Et5iXNLJYF5reTshdZ-wdAo7MKfVRvRPTU7Wz4TPTT4dLt6m_TL_3hc-54cofoREIIhcBQnhlgBk0xym42ggjVwemJcQAsYXaDBukS24OA82a1KdqQq5BA00YxkKzBnIFqr0eAf5floDVPFm
```

We now have retrieved the token and it can be used to authenticate and authorize a user.

## Creating Sample User for a Tenant

Execute the following command to create a tenant:

```
kubectl create tenant tenant-1
```

Simply add `--tenant tenant-1` to all commands in previous guide to create a User for a tenant, for example:

```
kubectl apply -f service-account.yaml --tenant tenant-1
```
