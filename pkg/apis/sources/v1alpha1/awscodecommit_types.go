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

// AWSCodeCommitSource is the Schema for an AWS CodeCommit event source.
type AWSCodeCommitSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSCodeCommitSourceSpec   `json:"spec,omitempty"`
	Status AWSCodeCommitSourceStatus `json:"status,omitempty"`
}

// Check the interfaces AWSCodeCommitSource should be implementing.
var (
	_ runtime.Object     = (*AWSCodeCommitSource)(nil)
	_ kmeta.OwnerRefable = (*AWSCodeCommitSource)(nil)
	_ apis.Validatable   = (*AWSCodeCommitSource)(nil)
	_ apis.Defaultable   = (*AWSCodeCommitSource)(nil)
	_ apis.HasSpec       = (*AWSCodeCommitSource)(nil)
)

// AWSCodeCommitSourceSpec defines the desired state of AWSCodeCommitSource.
type AWSCodeCommitSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// Repository is the name of the repository.
	Repository string `json:"repository"`
	// Branch is the name of the Git branch this source observes.
	Branch string `json:"branch"`
	// Region is the name of the AWS region where the repository is located.
	Region string `json:"region"`

	// Events is a list of events that should be processed by the source.
	Events []string `json:"events"`

	// Credentials to interact with the AWS CodeCommit API.
	Credentials AWSSecurityCredentials `json:"credentials"`
}

// AWSCodeCommitSourceStatus defines the observed state of AWSCodeCommitSource.
type AWSCodeCommitSourceStatus struct {
	duckv1.SourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSCodeCommitSourceList contains a list of AWSCodeCommitSource
type AWSCodeCommitSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSCodeCommitSource `json:"items"`
}
