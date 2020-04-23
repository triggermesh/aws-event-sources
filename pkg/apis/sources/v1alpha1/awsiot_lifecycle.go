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
func (s *AWSIoTSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSIoTSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSIoTSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// AWSIoTEventSource returns a representation of the source suitable for
// usage as a CloudEvent source.
func AWSIoTEventSource(endpoint, topic string) string {
	return fmt.Sprintf("%s/%s", endpoint, topic)
}

// AWSIoTEventType returns the given event type in a format suitable for
// usage as a CloudEvent type.
func AWSIoTEventType(eventType string) string {
	return fmt.Sprintf("com.amazon.iot.%s", eventType)
}
