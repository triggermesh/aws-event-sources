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

// AWSKinesisSource is the Schema for the event source.
type AWSKinesisSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSKinesisSourceSpec   `json:"spec,omitempty"`
	Status AWSKinesisSourceStatus `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object     = (*AWSKinesisSource)(nil)
	_ kmeta.OwnerRefable = (*AWSKinesisSource)(nil)
	_ apis.Validatable   = (*AWSKinesisSource)(nil)
	_ apis.Defaultable   = (*AWSKinesisSource)(nil)
	_ apis.HasSpec       = (*AWSKinesisSource)(nil)
)

// AWSKinesisSourceSpec defines the desired state of the event source.
type AWSKinesisSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// Name of the AWS Kinesis data stream.
	Stream string `json:"stream"`
	// Name of the AWS region where the AWS Kinesis data stream is located.
	Region string `json:"region"`

	// Credentials to interact with the AWS Kinesis API.
	Credentials AWSSecurityCredentials `json:"credentials"`
}

// AWSKinesisSourceStatus defines the observed state of the event source.
type AWSKinesisSourceStatus struct {
	duckv1.SourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSKinesisSourceList contains a list of event sources.
type AWSKinesisSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSKinesisSource `json:"items"`
}
