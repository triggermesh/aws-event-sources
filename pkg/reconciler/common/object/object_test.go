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

package object

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

func TestInjection(t *testing.T) {
	t.Run("Inject expected type", func(t *testing.T) {
		obj := &v1alpha1.AWSCodeCommitSource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "injected",
			},
		}

		ctx := context.TODO()
		ctx = With(ctx, obj)

		assert.Same(t, obj, FromContext(ctx), "Context injection preserves pointer to object")
	})

	t.Run("Inject unexpected type", func(t *testing.T) {
		type someType struct{}

		ctx := context.TODO()
		ctx = context.WithValue(ctx, objectKey{}, &someType{})

		assert.Nil(t, FromContext(ctx), "Failed type assertion returns nil")
	})
}
