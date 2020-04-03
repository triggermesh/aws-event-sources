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
	knative.dev/eventing => knative.dev/eventing v0.13.5
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200306230727-a56a6ea3fa56 // release-0.13 branch
)

require (
	cloud.google.com/go v0.56.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	go.uber.org/zap v1.14.1
	golang.org/x/tools v0.0.0-20200403190813-44a64ad78b9b // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/api v0.21.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/api v0.16.8
	k8s.io/apiextensions-apiserver v0.16.4 // indirect
	k8s.io/apimachinery v0.16.8
	k8s.io/client-go v0.16.8
	k8s.io/code-generator v0.16.8
	k8s.io/gengo v0.0.0-20200205140755-e0e292d8aa12 // indirect
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
	knative.dev/eventing v0.13.5
	knative.dev/pkg v0.0.0-20200306230727-a56a6ea3fa56
	sigs.k8s.io/yaml v1.2.0 // indirect
)
