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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewContainer(t *testing.T) {
	cont := NewContainer(tName,
		Port("h2c", 8080),
		Image(tImg),
		EnvVar("TEST_ENV1", "val1"),
		Port("health", 8081),
		EnvVar("TEST_ENV2", "val2"),
		Probe("/health", "health"),
		EnvVarFromSecret("TEST_ENV3", "test-secret", "someKey"),
	)

	expectCont := &corev1.Container{
		Name:  tName,
		Image: tImg,
		Ports: []corev1.ContainerPort{{
			Name:          "h2c",
			ContainerPort: 8080,
		}, {
			Name:          "health",
			ContainerPort: 8081,
		}},
		Env: []corev1.EnvVar{{
			Name:  "TEST_ENV1",
			Value: "val1",
		}, {
			Name:  "TEST_ENV2",
			Value: "val2",
		}, {
			Name: "TEST_ENV3",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "someKey",
				},
			},
		}},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromString("health"),
				},
			},
		},
	}

	if d := cmp.Diff(expectCont, cont); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
