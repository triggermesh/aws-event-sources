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

// Package object allows the injection of API objects into a context.
package object

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

type objectKey struct{}

// With returns a copy of the parent context in which the value associated with
// the object key is the given object.
func With(ctx context.Context, o runtime.Object) context.Context {
	return context.WithValue(ctx, objectKey{}, o)
}

// FromContext returns the object stored in the context.
func FromContext(ctx context.Context) runtime.Object {
	if o, ok := ctx.Value(objectKey{}).(runtime.Object); ok {
		return o
	}
	return nil
}
