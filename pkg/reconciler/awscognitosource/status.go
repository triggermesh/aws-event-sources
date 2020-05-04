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

package awscognitosource

import (
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// createCloudEventAttributes returns the CloudEvent types supported by the
// source.
func createCloudEventAttributes(srcSpec *v1alpha1.AWSCognitoSourceSpec) []duckv1.CloudEventAttributes {
	return []duckv1.CloudEventAttributes{
		{
			Type:   v1alpha1.AWSCognitoEventType(v1alpha1.AWSCognitoGenericEventType),
			Source: v1alpha1.AWSCognitoEventSource(srcSpec.IdentityPoolID),
		},
	}
}
