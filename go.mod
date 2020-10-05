module github.com/kubernetes/dashboard

go 1.12

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/igm/sockjs-go v2.0.1+incompatible // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.4.1
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/text v0.3.2
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/igm/sockjs-go.v2 v2.0.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.2.2
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/heapster v1.5.4
	k8s.io/klog v1.0.0
)

replace k8s.io/client-go => github.com/centaurus-cloud/arktos/staging/src/k8s.io/client-go v0.0.0-20200925053813-94992457ec50

replace k8s.io/apimachinery => github.com/centaurus-cloud/arktos/staging/src/k8s.io/apimachinery v0.0.0-20200925053813-94992457ec50

replace k8s.io/apiserver => github.com/centaurus-cloud/arktos/staging/src/k8s.io/apiserver v0.0.0-20200925053813-94992457ec50

replace k8s.io/api => github.com/centaurus-cloud/arktos/staging/src/k8s.io/api v0.0.0-20200925053813-94992457ec50

replace k8s.io/component-base => github.com/centaurus-cloud/arktos/staging/src/k8s.io/component-base v0.0.0-20200925053813-94992457ec50

replace k8s.io/apiextensions-apiserver => github.com/centaurus-cloud/arktos/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200925053813-94992457ec50

replace k8s.io/code-generator => github.com/centaurus-cloud/arktos/staging/src/k8s.io/code-generator v0.0.0-20200925053813-94992457ec50

replace k8s.io/cloud-provider => github.com/centaurus-cloud/arktos/staging/src/k8s.io/cloud-provider v0.0.0-20200925053813-94992457ec50

replace k8s.io/cri-api => github.com/centaurus-cloud/arktos/staging/src/k8s.io/cri-api v0.0.0-20200925053813-94992457ec50

replace k8s.io/csi-translation-lib => github.com/centaurus-cloud/arktos/staging/src/k8s.io/csi-translation-lib v0.0.0-20200925053813-94992457ec50

replace k8s.io/kubelet v0.0.0 => github.com/centaurus-cloud/arktos/staging/src/k8s.io/kubelet v0.0.0-20200925053813-94992457ec50

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
