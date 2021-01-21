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

package awscloudwatchsource

import (
	"encoding/json"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

const (
	envRegion          = "AWS_REGION"
	envQueries         = "QUERIES"
	envPollingInterval = "POLLING_INTERVAL"
)

const defaultPollingInterval = 5 * time.Minute

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awscloudwatchsource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterDeploymentBuilder returns an AdapterDeploymentBuilderFunc for the
// given source object and adapter config.
func adapterDeploymentBuilder(src *v1alpha1.AWSCloudWatchSource, cfg *adapterConfig) common.AdapterDeploymentBuilderFunc {
	return func(sinkURI *apis.URL) *appsv1.Deployment {
		var queries string
		if len(src.Spec.MetricQueries) > 0 {
			q, _ := json.Marshal(src.Spec.MetricQueries)
			queries = string(q)
		}

		pollingInterval := defaultPollingInterval
		if f := src.Spec.PollingInterval; f != nil && time.Duration(*f).Nanoseconds() > 0 {
			pollingInterval = time.Duration(*f)
		}

		return common.NewAdapterDeployment(src, sinkURI,
			resource.Image(cfg.Image),

			resource.EnvVar(envRegion, src.Spec.Region),
			resource.EnvVar(envQueries, queries),
			resource.EnvVar(envPollingInterval, pollingInterval.String()),
			resource.EnvVars(common.MakeSecurityCredentialsEnvVars(src.Spec.Credentials)...),
			resource.EnvVar(common.EnvNamespace, src.GetNamespace()),
			resource.EnvVar(common.EnvName, src.GetName()),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}
