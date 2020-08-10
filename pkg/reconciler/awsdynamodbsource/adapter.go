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

package awsdynamodbsource

import (
	appsv1 "k8s.io/api/apps/v1"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awsdynamodbsource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterDeploymentBuilder returns an AdapterDeploymentBuilderFunc for the
// given source object and adapter config.
func adapterDeploymentBuilder(src *v1alpha1.AWSDynamoDBSource, cfg *adapterConfig) common.AdapterDeploymentBuilderFunc {
	adapterName := common.AdapterName(src)

	return func(sinkURI *apis.URL) *appsv1.Deployment {
		name := kmeta.ChildName(adapterName+"-", src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		return resource.NewDeployment(src.Namespace, name,
			resource.Controller(src),

			resource.Label(common.AppNameLabel, adapterName),
			resource.Label(common.AppInstanceLabel, src.Name),
			resource.Label(common.AppComponentLabel, common.AdapterComponent),
			resource.Label(common.AppPartOfLabel, common.PartOf),
			resource.Label(common.AppManagedByLabel, common.ManagedBy),

			resource.Selector(common.AppNameLabel, adapterName),
			resource.Selector(common.AppInstanceLabel, src.Name),
			resource.PodLabel(common.AppComponentLabel, common.AdapterComponent),
			resource.PodLabel(common.AppPartOfLabel, common.PartOf),
			resource.PodLabel(common.AppManagedByLabel, common.ManagedBy),

			resource.Image(cfg.Image),

			resource.EnvVar(common.EnvName, src.Name),
			resource.EnvVar(common.EnvNamespace, src.Namespace),
			resource.EnvVar(common.EnvSink, sinkURIStr),
			resource.EnvVar(common.EnvARN, src.Spec.ARN.String()),
			resource.EnvVarFromSecret(common.EnvAccessKeyID,
				src.Spec.Credentials.AccessKeyID.ValueFromSecret.Name,
				src.Spec.Credentials.AccessKeyID.ValueFromSecret.Key),
			resource.EnvVarFromSecret(common.EnvSecretAccessKey,
				src.Spec.Credentials.SecretAccessKey.ValueFromSecret.Name,
				src.Spec.Credentials.SecretAccessKey.ValueFromSecret.Key),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}
