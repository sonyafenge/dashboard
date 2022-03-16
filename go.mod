module github.com/CentaurusInfra/dashboard

go 1.15

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/igm/sockjs-go v2.0.1+incompatible // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/lib/pq v1.10.4
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/text v0.3.2
	gopkg.in/igm/sockjs-go.v2 v2.0.0
	gopkg.in/square/go-jose.v2 v2.2.2
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/heapster v1.5.4
	k8s.io/klog v1.0.0
)

replace k8s.io/client-go => github.com/CentaurusInfra/arktos/staging/src/k8s.io/client-go v0.0.0-20200925053813-94992457ec50

replace k8s.io/apimachinery => github.com/CentaurusInfra/arktos/staging/src/k8s.io/apimachinery v0.0.0-20200925053813-94992457ec50

replace k8s.io/apiserver => github.com/CentaurusInfra/arktos/staging/src/k8s.io/apiserver v0.0.0-20200925053813-94992457ec50

replace k8s.io/api => github.com/CentaurusInfra/arktos/staging/src/k8s.io/api v0.0.0-20200925053813-94992457ec50

replace k8s.io/component-base => github.com/CentaurusInfra/arktos/staging/src/k8s.io/component-base v0.0.0-20200925053813-94992457ec50

replace k8s.io/apiextensions-apiserver => github.com/CentaurusInfra/arktos/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200925053813-94992457ec50

replace k8s.io/code-generator => github.com/CentaurusInfra/arktos/staging/src/k8s.io/code-generator v0.0.0-20200925053813-94992457ec50

replace k8s.io/cloud-provider => github.com/CentaurusInfra/arktos/staging/src/k8s.io/cloud-provider v0.0.0-20200925053813-94992457ec50

replace k8s.io/cri-api => github.com/CentaurusInfra/arktos/staging/src/k8s.io/cri-api v0.0.0-20200925053813-94992457ec50

replace k8s.io/csi-translation-lib => github.com/CentaurusInfra/arktos/staging/src/k8s.io/csi-translation-lib v0.0.0-20200925053813-94992457ec50

replace k8s.io/kubelet v0.0.0 => github.com/CentaurusInfra/arktos/staging/src/k8s.io/kubelet v0.0.0-20200925053813-94992457ec50

replace k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.0.0-20190718185405-0ce9869d0015

replace k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.0.0-20190718190146-f7b0473036f9

replace k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.0.0-20190718184434-a064d4d1ed7a

replace k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.0.0-20190718190030-ea930fedc880

replace k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.0.0-20190718185641-5233cb7cb41e

replace k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.0.0-20190718185913-d5429d807831

replace k8s.io/kubectl => k8s.io/kubectl v0.17.0

replace k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.0.0-20190718190548-039b99e58dbd

replace k8s.io/metrics v0.0.0 => k8s.io/metrics v0.0.0-20190718185242-1e1642704fe6

replace k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.0.0-20190718184639-baafa86838c0
