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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

const (
	// AWSCodeCommitConditionReady has status True when the AWSCodeCommitSource is ready to send events.
	AWSCodeCommitConditionReady = apis.ConditionReady
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (s *AWSCodeCommitSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("AWSCodeCommitSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *AWSCodeCommitSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *AWSCodeCommitSourceStatus) InitializeConditions() {}
