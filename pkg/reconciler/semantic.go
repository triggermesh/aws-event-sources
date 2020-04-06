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

package reconciler

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// Semantic can do semantic deep equality checks for Kubernetes API objects.
//
// For a given comparison function
//
//   comp(a, b interface{})
//
// 'a' should always be the desired state, and 'b' the current state for
// DeepDerivative comparisons to work as expected.
var Semantic = conversion.EqualitiesOrDie(
	deploymentEqual,
)

// eq is an instance of Equalities for internal derivative deep comparisons
// of API objects. Adapted from "k8s.io/apimachinery/equality".Semantic.
var eq = conversion.EqualitiesOrDie(
	func(a, b resource.Quantity) bool {
		if a.IsZero() {
			return true
		}
		return a.Cmp(b) == 0
	},
	func(a, b metav1.MicroTime) bool {
		if a.IsZero() {
			return true
		}
		return a.UTC() == b.UTC()
	},
	func(a, b metav1.Time) bool { // e.g. metadata.creationTimestamp
		if a.IsZero() {
			return true
		}
		return a.UTC() == b.UTC()
	},
	func(a, b labels.Selector) bool {
		if a.Empty() {
			return true
		}
		return a.String() == b.String()
	},
	func(a, b fields.Selector) bool {
		if a.Empty() {
			return true
		}
		return a.String() == b.String()
	},
	func(a, b int64) bool { // e.g. metadata.generation
		if a == 0 {
			return true
		}
		return a == b
	},
)

// deploymentEqual returns whether two Deployments are semantically equivalent.
func deploymentEqual(a, b *appsv1.Deployment) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepDerivative(&a.Spec, &b.Spec) {
		return false
	}

	return true
}
