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
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	pkgreconciler "knative.dev/pkg/reconciler"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/resource"
)

const adapterName = "awssnssource"

const (
	awsAccessKeyIdEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awssns"`
	// Logging configuration, serialized as JSON.
	LoggingCfg string `ignored:"true"`
	// Metrics (observability) configuration, serialized as JSON.
	MetricsCfg string `ignored:"true"`
}

// reconcileAdapter reconciles the state of the source's adapter.
func (r *Reconciler) reconcileAdapter(ctx context.Context,
	src *v1alpha1.AWSSNSSource) (*appsv1.Deployment, error) {

	sinkRef := &src.Spec.Sink.Ref
	if *sinkRef != nil && (*sinkRef).Namespace == "" {
		(*sinkRef).Namespace = src.Namespace
	}

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(src.Spec.Sink, src)
	if err != nil {
		reconciler.EventWarn(ctx, src, reconciler.ReasonBadSinkURI, "Could not resolve sink URI: %s", err)
		// skip adapter reconciliation if the sink URI can't be resolved.
		return nil, nil
	}
	desiredAdapter := makeAdapterDeployment(src, sinkURI.String(), r.adapterCfg)

	currentAdapter, err := r.getOrCreateAdapter(ctx, src, desiredAdapter)
	if err != nil {
		return nil, err
	}

	currentAdapter, err = r.syncAdapterDeployment(ctx, src, currentAdapter, desiredAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to synchronize adapter Deployment: %w", err)
	}

	return currentAdapter, nil
}

// getOrCreateAdapter returns the existing adapter Deployment for a given
// source, or creates it if it is missing.
func (r *Reconciler) getOrCreateAdapter(ctx context.Context, src *v1alpha1.AWSSNSSource,
	desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {

	adapter, err := r.deploymentLister(src.Namespace).Get(desiredAdapter.Name)
	switch {
	case apierrors.IsNotFound(err):
		adapter, err = r.deploymentClient(src.Namespace).Create(desiredAdapter)
		if err != nil {
			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedCreate,
				"Creation failed for adapter Deployment %q: %s", desiredAdapter.Name, err)
		}
		reconciler.Event(ctx, src, reconciler.ReasonCreate, "Created adapter Deployment %q", adapter.Name)

	case err != nil:
		return nil, fmt.Errorf("failed to get adapter Deployment from cache: %w", err)
	}

	return adapter, nil
}

// syncAdapterDeployment synchronizes the desired state of an adapter Deployment
// against its current state in the running cluster.
func (r *Reconciler) syncAdapterDeployment(ctx context.Context, src *v1alpha1.AWSSNSSource,
	currentAdapter, desiredAdapter *appsv1.Deployment) (*appsv1.Deployment, error) {

	if reconciler.Semantic.DeepEqual(desiredAdapter, currentAdapter) {
		return currentAdapter, nil
	}

	// resourceVersion must be returned to the API server unmodified for
	// optimistic concurrency, as per Kubernetes API conventions
	desiredAdapter.ResourceVersion = currentAdapter.ResourceVersion

	adapter, err := r.deploymentClient(currentAdapter.Namespace).Update(desiredAdapter)
	if err != nil {
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedUpdate,
			"Update failed for adapter Deployment %q: %s", desiredAdapter.Name, err)
	}
	reconciler.Event(ctx, src, reconciler.ReasonUpdate, "Updated adapter Deployment %q", adapter.Name)

	return adapter, nil
}

// makeAdapterDeployment returns a Deployment object for the source's adapter.
func makeAdapterDeployment(src *v1alpha1.AWSSNSSource, sinkURI string,
	adapterCfg *adapterConfig) *appsv1.Deployment {

	name := kmeta.ChildName(fmt.Sprintf("%s-", adapterName), src.Name)

	return resource.NewDeployment(src.Namespace, name,
		resource.Controller(src),

		resource.Label(reconciler.AppNameLabel, adapterName),
		resource.Label(reconciler.AppInstanceLabel, src.Name),
		resource.Label(reconciler.AppComponentLabel, reconciler.AdapterComponent),
		resource.Label(reconciler.AppPartOfLabel, reconciler.SourcesGroup),
		resource.Label(reconciler.AppManagedByLabel, reconciler.ManagedBy),

		resource.Selector(reconciler.AppNameLabel, adapterName),
		resource.Selector(reconciler.AppInstanceLabel, src.Name),
		resource.PodLabel(reconciler.AppComponentLabel, reconciler.AdapterComponent),
		resource.PodLabel(reconciler.AppPartOfLabel, reconciler.SourcesGroup),
		resource.PodLabel(reconciler.AppManagedByLabel, reconciler.ManagedBy),

		resource.Image(adapterCfg.Image),

		resource.EnvVar(reconciler.NameEnvVar, src.Name),
		resource.EnvVar(reconciler.NamespaceEnvVar, src.Namespace),
		resource.EnvVar(reconciler.SinkEnvVar, sinkURI),
		resource.EnvVar(reconciler.LoggingConfigEnvVar, adapterCfg.LoggingCfg),
		resource.EnvVar(reconciler.MetricsConfigEnvVar, adapterCfg.MetricsCfg),

		// TODO(antoineco): add source specific env vars
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
