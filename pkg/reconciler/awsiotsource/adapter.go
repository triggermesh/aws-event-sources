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

package awsiotsource

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

const (
	envEndpoint        = "THING_SHADOW_ENDPOINT"
	envTopic           = "THING_TOPIC"
	envRootCA          = "ROOT_CA"
	envRootCAPath      = "ROOT_CA_PATH"
	envCertificate     = "CERTIFICATE"
	envCertificatePath = "CERTIFICATE_PATH"
	envPrivateKey      = "PRIVATE_KEY"
	envPrivateKeyPath  = "PRIVATE_KEY_PATH"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awsiotsource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// Verify that Reconciler implements common.AdapterDeploymentBuilder.
var _ common.AdapterDeploymentBuilder = (*Reconciler)(nil)

// BuildAdapter implements common.AdapterDeploymentBuilder.
func (r *Reconciler) BuildAdapter(src v1alpha1.EventSource, sinkURI *apis.URL) *appsv1.Deployment {
	typedSrc := src.(*v1alpha1.AWSIoTSource)

	return common.NewAdapterDeployment(src, sinkURI,
		resource.Image(r.adapterCfg.Image),

		resource.EnvVar(envEndpoint, typedSrc.Spec.Endpoint),
		resource.EnvVar(envTopic, strings.TrimPrefix(typedSrc.Spec.ARN.Resource, "topic/")),
		resource.EnvVarFromSecret(envRootCA,
			typedSrc.Spec.RootCA.ValueFromSecret.Name,
			typedSrc.Spec.RootCA.ValueFromSecret.Key),
		resource.EnvVar(envRootCAPath, *typedSrc.Spec.RootCAPath),
		resource.EnvVarFromSecret(envCertificate,
			typedSrc.Spec.Certificate.ValueFromSecret.Name,
			typedSrc.Spec.Certificate.ValueFromSecret.Key),
		resource.EnvVar(envCertificatePath, *typedSrc.Spec.CertificatePath),
		resource.EnvVarFromSecret(envPrivateKey,
			typedSrc.Spec.PrivateKey.ValueFromSecret.Name,
			typedSrc.Spec.PrivateKey.ValueFromSecret.Key),
		resource.EnvVar(envPrivateKeyPath, *typedSrc.Spec.PrivateKeyPath),
		resource.EnvVars(r.adapterCfg.configs.ToEnvVars()...),
	)
}

// RBACOwners implements common.AdapterDeploymentBuilder.
func (r *Reconciler) RBACOwners(namespace string) ([]kmeta.OwnerRefable, error) {
	srcs, err := r.srcLister(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("listing objects from cache: %w", err)
	}

	ownerRefables := make([]kmeta.OwnerRefable, len(srcs))
	for i := range srcs {
		ownerRefables[i] = srcs[i]
	}

	return ownerRefables, nil
}
