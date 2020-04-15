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

package awsiotsource

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// syncStatus ensures the status of a given source is up-to-date.
func (r *Reconciler) syncStatus(src *v1alpha1.AWSIoTSource, adapter *appsv1.Deployment) error {
	currentStatus := &src.Status
	expectedStatus := r.computeStatus(src, adapter)

	if equality.Semantic.DeepEqual(expectedStatus, currentStatus) {
		return nil
	}

	src = &v1alpha1.AWSIoTSource{
		ObjectMeta: src.ObjectMeta,
		Status:     *expectedStatus,
	}

	_, err := r.sourceClient(src.Namespace).UpdateStatus(src)
	return err
}

// computeStatus returns the expected status of a given source.
func (r *Reconciler) computeStatus(src *v1alpha1.AWSIoTSource,
	adapter *appsv1.Deployment) *v1alpha1.AWSIoTSourceStatus {

	status := src.Status.DeepCopy()
	status.InitializeConditions()
	status.CloudEventAttributes = r.createCloudEventAttributes(&src.Spec)
	status.ObservedGeneration = src.Generation

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(src.Spec.Sink, src)
	if err != nil {
		status.MarkNoSink()
		return status
	}
	status.MarkSink(sinkURI)

	if adapter != nil {
		status.PropagateAvailability(adapter)
	}

	return status
}

// createCloudEventAttributes returns the CloudEvent types supported by the
// source.
func (r *Reconciler) createCloudEventAttributes(srcSpec *v1alpha1.AWSIoTSourceSpec) []duckv1.CloudEventAttributes {
	return []duckv1.CloudEventAttributes{
		{
			Type:   v1alpha1.AWSIoTEventType("greetings"),
			Source: v1alpha1.AWSIoTEventSource(srcSpec.Endpoint, srcSpec.Topic),
		},
	}
}
