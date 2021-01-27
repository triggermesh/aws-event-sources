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

package awssqssource

import (
	appsv1 "k8s.io/api/apps/v1"
	kr "k8s.io/apimachinery/pkg/api/resource"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awssqssource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

const healthPortName = "health"

// adapterDeploymentBuilder returns an AdapterDeploymentBuilderFunc for the
// given source object and adapter config.
func adapterDeploymentBuilder(src *v1alpha1.AWSSQSSource, cfg *adapterConfig) common.AdapterDeploymentBuilderFunc {
	return func(sinkURI *apis.URL) *appsv1.Deployment {
		return common.NewAdapterDeployment(src, sinkURI,
			resource.Image(cfg.Image),

			resource.EnvVar(common.EnvARN, src.Spec.ARN.String()),
			resource.EnvVars(common.MakeSecurityCredentialsEnvVars(src.Spec.Credentials)...),
			resource.EnvVar(common.EnvNamespace, src.GetNamespace()),
			resource.EnvVar(common.EnvName, src.GetName()),
			resource.EnvVars(cfg.configs.ToEnvVars()...),

			resource.Port(healthPortName, 8080),
			resource.Port("metrics", 9090),

			resource.Probe("/health", healthPortName),

			// CPU throttling can be observed below a limit of 1,
			// although the CPU usage under load remains below 400m.
			resource.Requests(
				*kr.NewMilliQuantity(90, kr.DecimalSI),     // 90m
				*kr.NewQuantity(1024*1024*30, kr.BinarySI), // 30Mi
			),
			resource.Limits(
				*kr.NewMilliQuantity(1000, kr.DecimalSI),   // 1
				*kr.NewQuantity(1024*1024*45, kr.BinarySI), // 45Mi
			),
		)
	}
}
