/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// AWSEventSource is implemented by all AWS event source types.
type AWSEventSource interface {
	metav1.Object
	runtime.Object
	// GetSink returns the source's event sink.
	GetSink() *duckv1.Destination
	// GetARN returns the source's AWS ARN.
	GetARN() string
	// GetStatus returns the source's status.
	GetStatus() *AWSEventSourceStatus
}

type sourceKey struct{}

// WithSource returns a copy of the parent context in which the value
// associated with the source key is the given event source.
func WithSource(ctx context.Context, s AWSEventSource) context.Context {
	return context.WithValue(ctx, sourceKey{}, s)
}

// SourceFromContext returns the source stored in the context.
func SourceFromContext(ctx context.Context) AWSEventSource {
	if s, ok := ctx.Value(sourceKey{}).(AWSEventSource); ok {
		return s
	}
	return nil
}
