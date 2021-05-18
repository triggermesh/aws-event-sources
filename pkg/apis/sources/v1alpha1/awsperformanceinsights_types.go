/*
Copyright (c) 2021 TriggerMesh Inc.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSPerformanceInsightsSource is the Schema for the event source.
type AWSPerformanceInsightsSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSPerformanceInsightsSourceSpec `json:"spec,omitempty"`
	Status EventSourceStatus                `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object = (*AWSPerformanceInsightsSource)(nil)
	_ EventSource    = (*AWSPerformanceInsightsSource)(nil)
)

// AWSPerformanceInsightsSourceSpec defines the desired state of the event source.
type AWSPerformanceInsightsSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	ARN apis.ARN `json:"arn"`
	// PollingInterval in a duration format for how often to pull metrics data from. Default is 5m
	// +optional
	PollingInterval *apis.Duration `json:"pollingInterval,omitempty"`

	Credentials AWSSecurityCredentials `json:"credentials"`

	MetricQuery string `json:"metricQuery"`

	Identifier string `json:"identifier"`

	ServiceType string `json:"serviceType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSPerformanceInsightsSourceList contains a list of event sources.
type AWSPerformanceInsightsSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSPerformanceInsightsSource `json:"items"`
}
