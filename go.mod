module github.com/triggermesh/aws-event-sources

go 1.15

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace k8s.io/client-go => k8s.io/client-go v0.22.5

require (
	github.com/aws/aws-sdk-go v1.37.1
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/emicklei/go-restful v2.15.0+incompatible // indirect
	github.com/google/go-cmp v0.5.7
	github.com/google/uuid v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.22.5
	knative.dev/eventing v0.30.0
	knative.dev/pkg v0.0.0-20220314170718-721abec0a377
	knative.dev/serving v0.30.0
)
