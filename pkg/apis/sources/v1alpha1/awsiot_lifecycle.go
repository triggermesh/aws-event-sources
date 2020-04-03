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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (s *AWSIoTSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSIoTSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSIoTSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// AWSIoTEventSource returns a representation of the source suitable for
// usage as a CloudEvent source.
func AWSIoTEventSource(endpoint, topic string) string {
	return fmt.Sprintf("%s/%s", endpoint, topic)
}

// AWSIoTEventType returns the given event type in a format suitable for
// usage as a CloudEvent type.
func AWSIoTEventType(eventType string) string {
	return fmt.Sprintf("com.amazon.iot.%s", eventType)
}

var awsIoTConditionSet = apis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *AWSIoTSourceStatus) InitializeConditions() {
	awsIoTConditionSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *AWSIoTSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if uri == nil {
		awsIoTConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
			ReasonSinkEmpty, "The sink has no URI")
		return
	}
	awsIoTConditionSet.Manage(s).MarkTrue(ConditionSinkProvided)
}

// MarkNoSink sets the SinkProvided condition to False.
func (s *AWSIoTSourceStatus) MarkNoSink() {
	s.SinkURI = nil
	awsIoTConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
		ReasonSinkNotFound, "The sink does not exist or its URI is not set")
}

// PropagateAvailability uses the readiness of the provided Deployment to
// determine whether the Deployed condition should be marked as true or false.
func (s *AWSIoTSourceStatus) PropagateAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		awsIoTConditionSet.Manage(s).MarkTrue(ConditionDeployed)
		return
	}

	msg := "The adapter Deployment is unavailable"

	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Message != "" {
			msg += ": " + cond.Message
		}
	}

	awsIoTConditionSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, msg)
}
