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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewContainer creates a Container object.
func NewContainer(name string, opts ...ObjectOption) *corev1.Container {
	c := &corev1.Container{
		Name: name,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Image sets a Container's image.
func Image(img string) ObjectOption {
	return func(object interface{}) {
		var image *string

		switch o := object.(type) {
		case *appsv1.Deployment:
			image = &firstContainer(o).Image
		case *corev1.Container:
			image = &o.Image
		}

		*image = img
	}
}

// Port adds a port to a Container.
func Port(name string, port int32) ObjectOption {
	return func(object interface{}) {
		var ports *[]corev1.ContainerPort

		switch o := object.(type) {
		case *corev1.Container:
			ports = &o.Ports
		case *appsv1.Deployment:
			ports = &firstContainer(o).Ports
		}

		*ports = append(*ports, corev1.ContainerPort{
			Name:          name,
			ContainerPort: port,
		})
	}
}

// EnvVar sets the value of a Container's environment variable.
func EnvVar(name, val string) ObjectOption {
	return func(object interface{}) {
		setEnvVar(envVarsFrom(object), name, val, nil)
	}
}

// EnvVars sets the value of multiple environment variables.
func EnvVars(evs ...corev1.EnvVar) ObjectOption {
	return func(object interface{}) {
		objEnvVars := envVarsFrom(object)
		*objEnvVars = append(*objEnvVars, evs...)
	}
}

// EnvVarFromSecret sets the value of a Container's environment variable to a
// reference to a Kubernetes Secret.
func EnvVarFromSecret(name, secretName, secretKey string) ObjectOption {
	return func(object interface{}) {
		valueFrom := &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		}

		setEnvVar(envVarsFrom(object), name, "", valueFrom)
	}
}

func envVarsFrom(object interface{}) (envVars *[]corev1.EnvVar) {
	switch o := object.(type) {
	case *corev1.Container:
		envVars = &o.Env
	case *appsv1.Deployment:
		envVars = &firstContainer(o).Env
	}

	return
}

func setEnvVar(envVars *[]corev1.EnvVar, name, value string, valueFrom *corev1.EnvVarSource) {
	*envVars = append(*envVars, corev1.EnvVar{
		Name:      name,
		Value:     value,
		ValueFrom: valueFrom,
	})
}

// Probe sets the HTTP readiness probe of a Deployment's first container.
func Probe(path, port string) ObjectOption {
	return func(object interface{}) {
		var rp **corev1.Probe

		switch o := object.(type) {
		case *corev1.Container:
			rp = &o.ReadinessProbe
		case *appsv1.Deployment:
			rp = &firstContainer(o).ReadinessProbe
		}

		*rp = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: path,
					Port: intstr.FromString(port),
				},
			},
		}
	}
}

// firstContainer returns a Deployment's first Container definition.
// A new empty Container is injected if the Deployment does not contain any.
func firstContainer(d *appsv1.Deployment) *corev1.Container {
	containers := &d.Spec.Template.Spec.Containers
	if len(*containers) == 0 {
		*containers = make([]corev1.Container, 1)
	}
	return &(*containers)[0]
}
