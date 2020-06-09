module github.com/triggermesh/aws-event-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/api => k8s.io/api v0.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8
)

require (
	github.com/aws/aws-sdk-go v1.30.21
	github.com/cloudevents/sdk-go/v2 v2.0.0-RC3
	github.com/google/go-cmp v0.4.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.0
	k8s.io/apiextensions-apiserver v0.18.0 // indirect
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
	knative.dev/eventing v0.14.1-0.20200512233457-09a6267a48c8
	knative.dev/pkg v0.0.0-20200519155757-14eb3ae3a5a7
	knative.dev/serving v0.14.1
)
