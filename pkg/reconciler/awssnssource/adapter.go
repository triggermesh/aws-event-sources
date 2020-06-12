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

package awssnssource

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
)

const adapterName = "awssnssource"

const (
	envSubscriptionAttrs = "SUBSCRIPTION_ATTRIBUTES"
	envPublicURL         = "PUBLIC_URL"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awssnssource"`

	// Public domain used to expose Knative Services
	Domain string `envconfig:"KNATIVE_DOMAIN" default:"example.com"`
	// URL scheme used to expose public Knative Services
	Scheme string `envconfig:"KNATIVE_URL_SCHEME" default:"http"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterServiceBuilder returns an AdapterServiceBuilderFunc for the
// given source object and adapter config.
func adapterServiceBuilder(src *v1alpha1.AWSSNSSource, cfg *adapterConfig) common.AdapterServiceBuilderFunc {
	return func(arn arn.ARN, sinkURI *apis.URL) *servingv1.Service {
		name := kmeta.ChildName(fmt.Sprintf("%s-", adapterName), src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		var subsAttrsStr string
		if subsAttrsJSON, err := json.Marshal(src.Spec.SubscriptionAttributes); err == nil {
			subsAttrsStr = string(subsAttrsJSON)
		}

		/* FIXME(antoineco): we wouldn't need to know the service URL in
		   advance if we could reconcile SNS subscriptions from the controller.
		   Ref. https://github.com/triggermesh/aws-event-sources/issues/157
		*/
		adapterURL := fmt.Sprintf("%s://%s.%s.%s", cfg.Scheme, name, src.Namespace, cfg.Domain)

		return resource.NewKnService(src.Namespace, name,
			resource.Controller(src),

			resource.Label(common.AppNameLabel, adapterName),
			resource.Label(common.AppInstanceLabel, src.Name),
			resource.Label(common.AppComponentLabel, common.AdapterComponent),
			resource.Label(common.AppPartOfLabel, common.PartOf),
			resource.Label(common.AppManagedByLabel, common.ManagedBy),

			resource.PodLabel(common.AppNameLabel, adapterName),
			resource.PodLabel(common.AppInstanceLabel, src.Name),
			resource.PodLabel(common.AppComponentLabel, common.AdapterComponent),
			resource.PodLabel(common.AppPartOfLabel, common.PartOf),
			resource.PodLabel(common.AppManagedByLabel, common.ManagedBy),

			resource.Image(cfg.Image),

			resource.EnvVar(common.EnvName, src.Name),
			resource.EnvVar(common.EnvNamespace, src.Namespace),
			resource.EnvVar(common.EnvSink, sinkURIStr),
			resource.EnvVar(common.EnvARN, arn.String()),
			resource.EnvVar(envSubscriptionAttrs, subsAttrsStr),
			resource.EnvVar(envPublicURL, adapterURL),
			resource.EnvVarFromSecret(common.EnvAccessKeyID,
				src.Spec.Credentials.AccessKeyID.ValueFromSecret.Name,
				src.Spec.Credentials.AccessKeyID.ValueFromSecret.Key),
			resource.EnvVarFromSecret(common.EnvSecretAccessKey,
				src.Spec.Credentials.SecretAccessKey.ValueFromSecret.Name,
				src.Spec.Credentials.SecretAccessKey.ValueFromSecret.Key),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
			/* FIXME(antoineco): default metrics port 9090 overlaps with queue-proxy
			 */
			resource.EnvVar(source.EnvMetricsCfg, ""),
		)
	}
}
