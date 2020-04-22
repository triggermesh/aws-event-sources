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

package awsdynamodbsource

import (
	appsv1 "k8s.io/api/apps/v1"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// computeStatus sets the attributes and conditions of the sources's status.
func (r *Reconciler) computeStatus(src *v1alpha1.AWSDynamoDBSource, adapter *appsv1.Deployment) {
	src.Status.InitializeConditions()
	src.Status.CloudEventAttributes = createCloudEventAttributes(&src.Spec)
	src.Status.ObservedGeneration = src.Generation

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(src.Spec.Sink, src)
	if err != nil {
		src.Status.MarkNoSink()
		return
	}
	src.Status.MarkSink(sinkURI)

	if adapter != nil {
		src.Status.PropagateAvailability(adapter)
	}
}

// createCloudEventAttributes returns the CloudEvent types supported by the
// source.
func createCloudEventAttributes(srcSpec *v1alpha1.AWSDynamoDBSourceSpec) []duckv1.CloudEventAttributes {
	types := v1alpha1.AWSDynamoDBEventTypes()
	ceAttributes := make([]duckv1.CloudEventAttributes, len(types))
	for i, typ := range types {
		ceAttributes[i] = duckv1.CloudEventAttributes{
			Type:   v1alpha1.AWSDynamoDBEventType(typ),
			Source: v1alpha1.AWSDynamoDBEventSource(srcSpec.Region, srcSpec.Table),
		}
	}
	return ceAttributes
}
