module github.com/triggermesh/aws-event-sources

go 1.14

require (
	github.com/aws/aws-sdk-go v1.31.12
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/go-cmp v0.5.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.18.7-rc.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/component-base v0.17.11
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.14.7
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	knative.dev/eventing v0.17.0
	knative.dev/pkg v0.0.0-20200812224206-44c860147a87
	knative.dev/serving v0.17.0
)

// Transitive dependencies of Knative.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)

// Kubernetes packages used by the e2e testing framework.
replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.11
	k8s.io/apiserver => k8s.io/apiserver v0.17.11
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.11
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.11
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.11
	k8s.io/component-base => k8s.io/component-base v0.17.11
	k8s.io/cri-api => k8s.io/cri-api v0.17.11
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.11
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.11
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.11
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.11
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.11
	k8s.io/kubectl => k8s.io/kubectl v0.17.11
	k8s.io/kubelet => k8s.io/kubelet v0.17.11
	k8s.io/kubernetes => k8s.io/kubernetes v1.17.11
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.11
	k8s.io/metrics => k8s.io/metrics v0.17.11
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.11
)

// Transitive dependencies of Kubernetes 1.17 (e2e testing framework).
replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.0.2
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/google/cadvisor => github.com/google/cadvisor v0.35.0
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	google.golang.org/api => google.golang.org/api v0.6.1-0.20190607001116-5213b8090861
	google.golang.org/grpc => google.golang.org/grpc v1.23.1
)
