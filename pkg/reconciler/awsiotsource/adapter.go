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

package awsiotsource

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/event"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/object"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/semantic"
)

const adapterName = "awsiotsource"

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
	// Logging configuration, serialized as JSON.
	LoggingCfg string `ignored:"true"`
	// Metrics (observability) configuration, serialized as JSON.
	MetricsCfg string `ignored:"true"`
}

// reconcileAdapter reconciles the state of the source's adapter.
func (r *Reconciler) reconcileAdapter(ctx context.Context) error {
	o := object.FromContext(ctx).(*v1alpha1.AWSIoTSource)

	sinkRef := &o.Spec.Sink.Ref
	if *sinkRef != nil && (*sinkRef).Namespace == "" {
		(*sinkRef).Namespace = o.Namespace
	}

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(o.Spec.Sink, o)
	if err != nil {
		o.Status.MarkNoSink()
		event.Warn(ctx, common.ReasonBadSinkURI, "Could not resolve sink URI: %s", err)
		// skip adapter reconciliation if the sink URI can't be resolved.
		return nil
	}
	o.Status.MarkSink(sinkURI)

	desiredAdapter := makeAdapterDeployment(ctx, sinkURI, r.adapterCfg)

	currentAdapter, err := r.getOrCreateAdapter(ctx, desiredAdapter)
	if err != nil {
		o.Status.PropagateAvailability(currentAdapter)
		return err
	}

	currentAdapter, err = r.syncAdapterDeployment(ctx, currentAdapter, desiredAdapter)
	if err != nil {
		return fmt.Errorf("failed to synchronize adapter Deployment: %w", err)
	}
	o.Status.PropagateAvailability(currentAdapter)

	return nil
}

// getOrCreateAdapter returns the existing adapter Deployment for a given
// source, or creates it if it is missing.
func (r *Reconciler) getOrCreateAdapter(ctx context.Context, desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {
	o := object.FromContext(ctx).(*v1alpha1.AWSIoTSource)

	adapter, err := r.deploymentLister(o.Namespace).Get(desiredAdapter.Name)
	switch {
	case apierrors.IsNotFound(err):
		adapter, err = r.deploymentClient(o.Namespace).Create(desiredAdapter)
		if err != nil {
			return nil, reconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedAdapterCreate,
				"Failed to create adapter Deployment %q: %s", desiredAdapter.Name, err)
		}
		event.Normal(ctx, common.ReasonAdapterCreate, "Created adapter Deployment %q", adapter.Name)

	case err != nil:
		return nil, fmt.Errorf("failed to get adapter Deployment from cache: %w", err)
	}

	return adapter, nil
}

// syncAdapterDeployment synchronizes the desired state of an adapter Deployment
// against its current state in the running cluster.
func (r *Reconciler) syncAdapterDeployment(ctx context.Context,
	currentAdapter, desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {

	if semantic.Semantic.DeepEqual(desiredAdapter, currentAdapter) {
		return currentAdapter, nil
	}

	// resourceVersion must be returned to the API server unmodified for
	// optimistic concurrency, as per Kubernetes API conventions
	desiredAdapter.ResourceVersion = currentAdapter.ResourceVersion

	// (fake Clientset) preserve status to avoid resetting conditions
	desiredAdapter.Status = currentAdapter.Status

	adapter, err := r.deploymentClient(currentAdapter.Namespace).Update(desiredAdapter)
	if err != nil {
		return nil, reconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedAdapterUpdate,
			"Failed to update adapter Deployment %q: %s", desiredAdapter.Name, err)
	}
	event.Normal(ctx, common.ReasonAdapterUpdate, "Updated adapter Deployment %q", adapter.Name)

	return adapter, nil
}

// makeAdapterDeployment returns a Deployment object for the source's adapter.
func makeAdapterDeployment(ctx context.Context, sinkURI *apis.URL, adapterCfg *adapterConfig) *appsv1.Deployment {
	o := object.FromContext(ctx).(*v1alpha1.AWSIoTSource)
	name := kmeta.ChildName(fmt.Sprintf("%s-", adapterName), o.Name)

	var sinkURIStr string
	if sinkURI != nil {
		sinkURIStr = sinkURI.String()
	}

	return resource.NewDeployment(o.Namespace, name,
		resource.Controller(o),

		resource.Label(common.AppNameLabel, adapterName),
		resource.Label(common.AppInstanceLabel, o.Name),
		resource.Label(common.AppComponentLabel, common.AdapterComponent),
		resource.Label(common.AppPartOfLabel, common.PartOf),
		resource.Label(common.AppManagedByLabel, common.ManagedBy),

		resource.Selector(common.AppNameLabel, adapterName),
		resource.Selector(common.AppInstanceLabel, o.Name),
		resource.PodLabel(common.AppComponentLabel, common.AdapterComponent),
		resource.PodLabel(common.AppPartOfLabel, common.PartOf),
		resource.PodLabel(common.AppManagedByLabel, common.ManagedBy),

		resource.Image(adapterCfg.Image),

		resource.EnvVar(common.NameEnvVar, o.Name),
		resource.EnvVar(common.NamespaceEnvVar, o.Namespace),
		resource.EnvVar(common.SinkEnvVar, sinkURIStr),
		resource.EnvVar(common.LoggingConfigEnvVar, adapterCfg.LoggingCfg),
		resource.EnvVar(common.MetricsConfigEnvVar, adapterCfg.MetricsCfg),
		resource.EnvVar(envEndpoint, o.Spec.Endpoint),
		resource.EnvVar(envTopic, o.Spec.Topic),
		resource.EnvVarFromSecret(envRootCA,
			o.Spec.RootCA.ValueFromSecret.Name,
			o.Spec.RootCA.ValueFromSecret.Key),
		resource.EnvVar(envRootCAPath, *o.Spec.RootCAPath),
		resource.EnvVarFromSecret(envCertificate,
			o.Spec.Certificate.ValueFromSecret.Name,
			o.Spec.Certificate.ValueFromSecret.Key),
		resource.EnvVar(envCertificatePath, *o.Spec.CertificatePath),
		resource.EnvVarFromSecret(envPrivateKey,
			o.Spec.PrivateKey.ValueFromSecret.Name,
			o.Spec.PrivateKey.ValueFromSecret.Key),
		resource.EnvVar(envPrivateKeyPath, *o.Spec.PrivateKeyPath),
	)
}

// updateAdapterLoggingConfig serializes the logging config from a ConfigMap to
// JSON and updates the existing config stored in the Reconciler.
func (r *Reconciler) updateAdapterLoggingConfig(cfg *corev1.ConfigMap) {
	delete(cfg.Data, "_example")

	logCfg, err := logging.NewConfigFromConfigMap(cfg)
	if err != nil {
		r.logger.Warnw("Failed to create adapter logging config from ConfigMap",
			"configmap", cfg.Name, "error", err)
		return
	}

	logCfgJSON, err := logging.LoggingConfigToJson(logCfg)
	if err != nil {
		r.logger.Warnw("Failed to serialize adapter logging config to JSON",
			"configmap", cfg.Name, "error", err)
		return
	}

	r.adapterCfg.LoggingCfg = logCfgJSON

	r.logger.Infow("Updated adapter logging config from ConfigMap", "configmap", cfg.Name)
}

// updateAdapterMetricsConfig serializes the metrics config from a ConfigMap to
// JSON and updates the existing config stored in the Reconciler.
func (r *Reconciler) updateAdapterMetricsConfig(cfg *corev1.ConfigMap) {
	delete(cfg.Data, "_example")

	metricsCfg := &metrics.ExporterOptions{
		Domain:    metrics.Domain(),
		Component: adapterName,
		ConfigMap: cfg.Data,
	}

	metricsCfgJSON, err := metrics.MetricsOptionsToJson(metricsCfg)
	if err != nil {
		r.logger.Warnw("Failed to serialize adapter metrics config to JSON",
			"configmap", cfg.Name, "error", err)
		return
	}

	r.adapterCfg.MetricsCfg = metricsCfgJSON

	r.logger.Infow("Updated adapter metrics config from ConfigMap", "configmap", cfg.Name)
}
