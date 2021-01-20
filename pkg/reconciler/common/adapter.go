/*
Copyright (c) 2020-2021 TriggerMesh Inc.

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
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

const metricsPrometheusPort uint16 = 9092

// AdapterName returns the adapter's name for the given source object.
func AdapterName(o kmeta.OwnerRefable) string {
	return strings.ToLower(o.GetGroupVersionKind().Kind)
}

// NewAdapterDeployment is a wrapper around resource.NewDeployment which
// pre-populates attributes common to all adapter Deployments.
func NewAdapterDeployment(src kmeta.OwnerRefable, sinkURI *apis.URL, opts ...resource.ObjectOption) *appsv1.Deployment {
	app := AdapterName(src)
	meta := src.GetObjectMeta()
	srcNs := meta.GetNamespace()
	srcName := meta.GetName()

	var sinkURIStr string
	if sinkURI != nil {
		sinkURIStr = sinkURI.String()
	}

	return resource.NewDeployment(srcNs, kmeta.ChildName(app+"-", srcName),
		append([]resource.ObjectOption{
			resource.TerminationErrorToLogs,
			resource.Controller(src),

			resource.Label(appNameLabel, app),
			resource.Label(appInstanceLabel, srcName),
			resource.Label(appComponentLabel, componentAdapter),
			resource.Label(appPartOfLabel, partOf),
			resource.Label(appManagedByLabel, managedBy),

			resource.Selector(appNameLabel, app),
			resource.Selector(appInstanceLabel, srcName),
			resource.PodLabel(appComponentLabel, componentAdapter),
			resource.PodLabel(appPartOfLabel, partOf),
			resource.PodLabel(appManagedByLabel, managedBy),

			resource.EnvVar(envNamespace, srcNs),
			resource.EnvVar(envName, srcName),
			resource.EnvVar(envSink, sinkURIStr),
		}, opts...)...,
	)
}

// NewAdapterKnService is a wrapper around resource.NewKnService which
// pre-populates attributes common to all adapter Knative Services.
func NewAdapterKnService(src kmeta.OwnerRefable, sinkURI *apis.URL, opts ...resource.ObjectOption) *servingv1.Service {
	app := AdapterName(src)
	meta := src.GetObjectMeta()
	srcNs := meta.GetNamespace()
	srcName := meta.GetName()

	var sinkURIStr string
	if sinkURI != nil {
		sinkURIStr = sinkURI.String()
	}

	return resource.NewKnService(srcNs, kmeta.ChildName(app+"-", srcName),
		append([]resource.ObjectOption{
			resource.Controller(src),

			resource.Label(appNameLabel, app),
			resource.Label(appInstanceLabel, srcName),
			resource.Label(appComponentLabel, componentAdapter),
			resource.Label(appPartOfLabel, partOf),
			resource.Label(appManagedByLabel, managedBy),

			resource.PodLabel(appNameLabel, app),
			resource.PodLabel(appInstanceLabel, srcName),
			resource.PodLabel(appComponentLabel, componentAdapter),
			resource.PodLabel(appPartOfLabel, partOf),
			resource.PodLabel(appManagedByLabel, managedBy),

			resource.EnvVar(envNamespace, srcNs),
			resource.EnvVar(envName, srcName),
			resource.EnvVar(envSink, sinkURIStr),
			resource.EnvVar(envMetricsPrometheusPort, strconv.FormatUint(uint64(metricsPrometheusPort), 10)),
		}, opts...)...,
	)
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
