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

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (s *AWSCognitoSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSCognitoSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSCognitoSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// AWSCognitoEventSource returns a representation of the source suitable for
// usage as a CloudEvent source.
func AWSCognitoEventSource(identityPoolID string) string {
	return fmt.Sprintf("aws:cognito:identitypool/%s", identityPoolID)
}

// Supported event types
const (
	AWSCognitoGenericEventType = "sync_trigger"
)

// AWSCognitoEventType returns the given event type in a format suitable for
// usage as a CloudEvent type.
func AWSCognitoEventType(eventType string) string {
	return fmt.Sprintf("com.amazon.cognito.%s", eventType)
}
