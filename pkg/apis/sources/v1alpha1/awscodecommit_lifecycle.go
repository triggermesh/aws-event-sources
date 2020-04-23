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
