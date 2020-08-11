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

package common

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmeta"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// AdapterName returns the adapter's name for the given source object.
func AdapterName(o kmeta.OwnerRefable) string {
	return strings.ToLower(o.GetGroupVersionKind().Kind)
}

// MakeSecurityCredentialsEnvVars returns environment variables for the given
// AWS security credentials.
func MakeSecurityCredentialsEnvVars(creds v1alpha1.AWSSecurityCredentials) []corev1.EnvVar {
	const (
		envAccessKeyID = iota
		envSecretAccessKey
	)

	credsEnvVars := []corev1.EnvVar{
		{Name: EnvAccessKeyID},
		{Name: EnvSecretAccessKey},
	}

	if vfs := creds.AccessKeyID.ValueFromSecret; vfs != nil {
		credsEnvVars[envAccessKeyID].ValueFrom = envVarValueFromSecret(vfs.Name, vfs.Key)
	} else {
		credsEnvVars[envAccessKeyID].Value = creds.AccessKeyID.Value
	}

	if vfs := creds.SecretAccessKey.ValueFromSecret; vfs != nil {
		credsEnvVars[envSecretAccessKey].ValueFrom = envVarValueFromSecret(vfs.Name, vfs.Key)
	} else {
		credsEnvVars[envSecretAccessKey].Value = creds.SecretAccessKey.Value
	}

	return credsEnvVars
}

// envVarValueFromSecret returns the value of an environment variable sourced
// from a Kubernetes Secret.
func envVarValueFromSecret(secretName, secretKey string) *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
			Key: secretKey,
		},
	}
}
