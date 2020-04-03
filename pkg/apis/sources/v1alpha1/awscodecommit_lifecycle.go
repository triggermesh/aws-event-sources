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
func (s *AWSCodeCommitSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSCodeCommitSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSCodeCommitSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// AWSCodeCommitEventSource returns a representation of the source suitable for
// usage as a CloudEvent source.
func AWSCodeCommitEventSource(region, repo string) string {
	return fmt.Sprintf("https://git-codecommit.%s.amazonaws.com/v1/repos/%s", region, repo)
}

// AWSCodeCommitEventType returns the given event type in a format suitable for
// usage as a CloudEvent type.
func AWSCodeCommitEventType(eventType string) string {
	return fmt.Sprintf("com.amazon.codecommit.%s", eventType)
}

var awsCodeCommitConditionSet = apis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *AWSCodeCommitSourceStatus) InitializeConditions() {
	awsCodeCommitConditionSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *AWSCodeCommitSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if uri == nil {
		awsCodeCommitConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
			ReasonSinkEmpty, "The sink has no URI")
		return
	}
	awsCodeCommitConditionSet.Manage(s).MarkTrue(ConditionSinkProvided)
}

// MarkNoSink sets the SinkProvided condition to False.
func (s *AWSCodeCommitSourceStatus) MarkNoSink() {
	s.SinkURI = nil
	awsCodeCommitConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
		ReasonSinkNotFound, "The sink does not exist or its URI is not set")
}

// PropagateAvailability uses the readiness of the provided Deployment to
// determine whether the Deployed condition should be marked as true or false.
func (s *AWSCodeCommitSourceStatus) PropagateAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		awsCodeCommitConditionSet.Manage(s).MarkTrue(ConditionDeployed)
		return
	}

	msg := "The adapter Deployment is unavailable"

	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Message != "" {
			msg += ": " + cond.Message
		}
	}

	awsCodeCommitConditionSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, msg)
}
