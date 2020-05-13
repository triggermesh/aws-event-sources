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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/event"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/object"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/resource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/semantic"
)

const adapterName = "awsdynamodbsource"

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/awsdynamodbsource"`
	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// reconcileAdapter reconciles the state of the source's adapter.
func (r *Reconciler) reconcileAdapter(ctx context.Context, arn arn.ARN) error {
	o := object.FromContext(ctx).(*v1alpha1.AWSDynamoDBSource)

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

	desiredAdapter := makeAdapterDeployment(ctx, arn, sinkURI, r.adapterCfg)

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
	o := object.FromContext(ctx).(*v1alpha1.AWSDynamoDBSource)

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
func makeAdapterDeployment(ctx context.Context, arn arn.ARN,
	sinkURI *apis.URL, adapterCfg *adapterConfig) *appsv1.Deployment {

	o := object.FromContext(ctx).(*v1alpha1.AWSDynamoDBSource)
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

		resource.EnvVar(common.EnvName, o.Name),
		resource.EnvVar(common.EnvNamespace, o.Namespace),
		resource.EnvVar(common.EnvSink, sinkURIStr),
		resource.EnvVar(common.EnvARN, arn.String()),
		resource.EnvVarFromSecret(common.EnvAccessKeyID,
			o.Spec.Credentials.AccessKeyID.ValueFromSecret.Name,
			o.Spec.Credentials.AccessKeyID.ValueFromSecret.Key),
		resource.EnvVarFromSecret(common.EnvSecretAccessKey,
			o.Spec.Credentials.SecretAccessKey.ValueFromSecret.Name,
			o.Spec.Credentials.SecretAccessKey.ValueFromSecret.Key),
		resource.EnvVars(adapterCfg.configs.ToEnvVars()...),
	)
}
