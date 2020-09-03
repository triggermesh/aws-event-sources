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

// Package conversion contains helpers to convert source types to/from Unstructured.
package conversion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	internalscheme "github.com/triggermesh/aws-event-sources/pkg/client/generated/clientset/internalclientset/scheme"
)

func init() {
	sb := runtime.NewSchemeBuilder(
		internalscheme.AddToScheme,
	)
	if err := sb.AddToScheme(scheme.Scheme); err != nil {
		panic(fmt.Errorf("error adding internal types to Scheme: %s", err))
	}
}
