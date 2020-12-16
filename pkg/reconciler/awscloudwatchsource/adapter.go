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

package awscloudwatchsource

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/pkg/logging"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

const (
	envRegion          = "AWS_REGION"
	envQueries         = "QUERIES"
	envPollingInterval = "POLLING_INTERVAL"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awscloudwatchsource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
	// Context for logging
	context context.Context
}

// adapterDeploymentBuilder returns an AdapterDeploymentBuilderFunc for the
// given source object and adapter config.
func adapterDeploymentBuilder(src *v1alpha1.AWSCloudWatchSource, cfg *adapterConfig) common.AdapterDeploymentBuilderFunc {
	adapterName := common.AdapterName(src)

	return func(sinkURI *apis.URL) *appsv1.Deployment {
		name := kmeta.ChildName(adapterName+"-", src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		var queries string
		if src.Spec.MetricQueries != nil {
			q, err := json.Marshal(src.Spec.MetricQueries)
			if err != nil {
				logging.FromContext(cfg.context).Errorw("Unable to obtain metrics: ", zap.Error(err))
			}

			queries = string(q)
		}

		var pollingInterval string
		if src.Spec.PollingFrequency != nil {
			pollingInterval = *src.Spec.PollingFrequency
		} else {
			// Defaulting to 5m which should keep things within the free tier
			pollingInterval = "5m"
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
			resource.EnvVar(envRegion, src.Spec.Region),
			resource.EnvVar(envQueries, queries),
			resource.EnvVar(envPollingInterval, pollingInterval),
			resource.EnvVars(common.MakeSecurityCredentialsEnvVars(src.Spec.Credentials)...),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}
