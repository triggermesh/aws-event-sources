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

// AWSCognitoSource is the Schema for the event source.
type AWSCognitoSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSCognitoSourceSpec `json:"spec,omitempty"`
	Status AWSEventSourceStatus `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object     = (*AWSCognitoSource)(nil)
	_ kmeta.OwnerRefable = (*AWSCognitoSource)(nil)
	_ apis.Validatable   = (*AWSCognitoSource)(nil)
	_ apis.Defaultable   = (*AWSCognitoSource)(nil)
	_ apis.HasSpec       = (*AWSCognitoSource)(nil)
)

// AWSCognitoSourceSpec defines the desired state of the event source.
type AWSCognitoSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// ID number of the identity pool.
	IdentityPoolID string `json:"identityPoolID"`

	// Credentials to interact with the AWS Cognito API.
	Credentials AWSSecurityCredentials `json:"credentials"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSCognitoSourceList contains a list of event sources.
type AWSCognitoSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSCognitoSource `json:"items"`
}
