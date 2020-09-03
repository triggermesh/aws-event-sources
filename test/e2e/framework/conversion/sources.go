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

package conversion

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// ToUnstructured converts a source object to its Unstructured representation.
func ToUnstructured(src runtime.Object) *unstructured.Unstructured {
	fnUnstr := &unstructured.Unstructured{}

	convertCtx := runtime.NewMultiGroupVersioner(v1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(src, fnUnstr, convertCtx); err != nil {
		framework.Failf("Error converting source to Unstructured: %v", err)
	}

	return fnUnstr
}

// FromUnstructured converts an instance of Unstructured to a source object.
func FromUnstructured(fn *unstructured.Unstructured, srcObjPtr runtime.Object) {
	convertCtx := runtime.NewMultiGroupVersioner(v1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(fn, srcObjPtr, convertCtx); err != nil {
		framework.Failf("Error converting Unstructured to source: %v", err)
	}
}

// V1alpha1GVR returns a schema.GroupVersionResource for the given
// source schema.GroupResource.
func V1alpha1GVR(gr schema.GroupResource) schema.GroupVersionResource {
	return gr.WithVersion(v1alpha1.SchemeGroupVersion.Version)
}
