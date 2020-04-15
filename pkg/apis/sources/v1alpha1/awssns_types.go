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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSSNSSource is the Schema for the event source.
type AWSSNSSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSNSSourceSpec   `json:"spec,omitempty"`
	Status AWSSNSSourceStatus `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object     = (*AWSSNSSource)(nil)
	_ kmeta.OwnerRefable = (*AWSSNSSource)(nil)
	_ apis.Validatable   = (*AWSSNSSource)(nil)
	_ apis.Defaultable   = (*AWSSNSSource)(nil)
	_ apis.HasSpec       = (*AWSSNSSource)(nil)
)

// AWSSNSSourceSpec defines the desired state of the event source.
type AWSSNSSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// Name of the AWS SNS topic.
	Topic string `json:"topic"`
	// Name of the AWS region where the AWS SNS topic is located.
	Region string `json:"region"`

	// Credentials to interact with the AWS SNS API.
	Credentials AWSSecurityCredentials `json:"credentials"`
}

// AWSSNSSourceStatus defines the observed state of the event source.
type AWSSNSSourceStatus struct {
	duckv1.SourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSSNSSourceList contains a list of event sources.
type AWSSNSSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSNSSource `json:"items"`
}
