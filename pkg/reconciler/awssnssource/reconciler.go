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

	"go.uber.org/zap"

	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"

	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awssnssource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/object"
)

// Reconciler implements controller.Reconciler for the event source type.
type Reconciler struct {
	logger *zap.SugaredLogger

	// URI resolver for sinks
	sinkResolver *resolver.URIResolver

	// adapter properties
	adapterCfg *adapterConfig

	// API clients
	deploymentClient func(namespace string) appsclientv1.DeploymentInterface

	// objects listers
	deploymentLister func(namespace string) appslistersv1.DeploymentNamespaceLister
}

// Check that our Reconciler implements Interface
var _ reconcilerv1alpha1.Interface = (*Reconciler)(nil)

// Optionally check that our Reconciler implements Finalizer
//var _ reconcilerv1alpha1.Finalizer = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *v1alpha1.AWSSNSSource) reconciler.Event {
	// inject object into context for usage in event recorder and
	// reconciliation logic
	ctx = object.With(ctx, o)

	o.Status.InitializeConditions()
	o.Status.ObservedGeneration = o.Generation
	o.Status.CloudEventAttributes = createCloudEventAttributes(&o.Spec)

	return r.reconcileAdapter(ctx)
}

// Optionally, use FinalizeKind to add finalizers. FinalizeKind will be called
// when the resource is deleted.
//func (r *Reconciler) FinalizeKind(ctx context.Context, o *v1alpha1.AWSSNSSource) reconciler.Event {
//	// TODO: add custom finalization logic here.
//	return nil
//}
