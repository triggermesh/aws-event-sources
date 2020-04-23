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

package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	. "github.com/triggermesh/aws-event-sources/pkg/reconciler/testing"
)

const (
	tNs   = "testns"
	tName = "testname"

	tImg = "registry/image:tag"
)

func TestMetaObjectOptions(t *testing.T) {
	objMeta := NewDeployment(tNs, tName,
		Label("test.label/2", "val2"),
		Controller(DummyOwnerRefable()),
		Label("test.label/1", "val1"),
	).ObjectMeta

	expectObjMeta := metav1.ObjectMeta{
		Namespace: tNs,
		Name:      tName,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(DummyOwnerRefable()),
		},
		Labels: map[string]string{
			"test.label/1": "val1",
			"test.label/2": "val2",
		},
	}

	if d := cmp.Diff(expectObjMeta, objMeta); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
