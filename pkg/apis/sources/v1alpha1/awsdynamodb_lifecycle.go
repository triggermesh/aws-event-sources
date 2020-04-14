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
func (s *AWSDynamoDBSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSDynamoDBSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSDynamoDBSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// AWSDynamoDBEventSource returns a representation of the source suitable for
// usage as a CloudEvent source.
func AWSDynamoDBEventSource(region, table string) string {
	return fmt.Sprintf("%s:table/%s", region, table)
}

// Supported event types
const (
	AWSDynamoDBAddEventType    = "insert"
	AWSDynamoDBModifyEventType = "modify"
	AWSDynamoDBRemoveEventType = "remove"
)

// AWSDynamoDBEventTypes returns the list of event types supported by the event
// source.
func AWSDynamoDBEventTypes() []string {
	return []string{
		AWSDynamoDBAddEventType,
		AWSDynamoDBModifyEventType,
		AWSDynamoDBRemoveEventType,
	}
}

// AWSDynamoDBEventType returns the given event type in a format suitable for
// usage as a CloudEvent type.
func AWSDynamoDBEventType(eventType string) string {
	return fmt.Sprintf("com.amazon.dynamodb.%s", eventType)
}

var awsDynamoDBConditionSet = apis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *AWSDynamoDBSourceStatus) InitializeConditions() {
	awsDynamoDBConditionSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *AWSDynamoDBSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if uri == nil {
		awsDynamoDBConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
			ReasonSinkEmpty, "The sink has no URI")
		return
	}
	awsDynamoDBConditionSet.Manage(s).MarkTrue(ConditionSinkProvided)
}

// MarkNoSink sets the SinkProvided condition to False.
func (s *AWSDynamoDBSourceStatus) MarkNoSink() {
	s.SinkURI = nil
	awsDynamoDBConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
		ReasonSinkNotFound, "The sink does not exist or its URI is not set")
}

// PropagateAvailability uses the readiness of the provided Deployment to
// determine whether the Deployed condition should be marked as true or false.
func (s *AWSDynamoDBSourceStatus) PropagateAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		awsDynamoDBConditionSet.Manage(s).MarkTrue(ConditionDeployed)
		return
	}

	msg := "The adapter Deployment is unavailable"

	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Message != "" {
			msg += ": " + cond.Message
		}
	}

	awsDynamoDBConditionSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, msg)
}
