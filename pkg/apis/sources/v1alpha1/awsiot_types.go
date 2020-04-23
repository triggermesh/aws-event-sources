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

// AWSIoTSource is the Schema for the event source.
type AWSIoTSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIoTSourceSpec     `json:"spec,omitempty"`
	Status AWSEventSourceStatus `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object     = (*AWSIoTSource)(nil)
	_ kmeta.OwnerRefable = (*AWSIoTSource)(nil)
	_ apis.Validatable   = (*AWSIoTSource)(nil)
	_ apis.Defaultable   = (*AWSIoTSource)(nil)
	_ apis.HasSpec       = (*AWSIoTSource)(nil)
)

// AWSIoTSourceSpec defines the desired state of the event source.
type AWSIoTSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// Host name of the endpoint the client connects to
	Endpoint string `json:"endpoint"`
	// Topic messages get published to
	Topic string `json:"topic"`

	// Contents of the root CA
	RootCA ValueFromField `json:"rootCA"`
	// Path where the root CA gets written
	// +optional
	RootCAPath *string `json:"rootCAPath,omitempty"`

	// Contents of the client certificate
	Certificate ValueFromField `json:"certificate"`
	// Path where the client certificate gets written
	// +optional
	CertificatePath *string `json:"certificatePath,omitempty"`

	// Contents of the client private key
	PrivateKey ValueFromField `json:"privateKey"`
	// Path where the client private key gets written
	// +optional
	PrivateKeyPath *string `json:"privateKeyPath,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSIoTSourceList contains a list of event sources.
type AWSIoTSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSIoTSource `json:"items"`
}
