module github.com/triggermesh/aws-event-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	k8s.io/api => k8s.io/api v0.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8
)

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	github.com/aws/aws-sdk-go v1.30.6 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	go.uber.org/zap v1.14.1
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/api v0.15.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/api v0.18.0
	k8s.io/apiextensions-apiserver v0.18.0 // indirect
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v0.18.0
	k8s.io/code-generator v0.18.0
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
	knative.dev/eventing v0.13.1-0.20200408164402-7babc039e52b
	knative.dev/pkg v0.0.0-20200407145900-0c36abbff9e5
)
