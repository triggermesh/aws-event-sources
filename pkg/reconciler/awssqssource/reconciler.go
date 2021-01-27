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
	"context"

	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awssqssource"
	listersv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/listers/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
)

// Reconciler implements controller.Reconciler for the event source type.
type Reconciler struct {
	base       common.GenericDeploymentReconciler
	adapterCfg *adapterConfig

	srcLister func(namespace string) listersv1alpha1.AWSSQSSourceNamespaceLister
}

// Check that our Reconciler implements Interface
var _ reconcilerv1alpha1.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.AWSSQSSource) reconciler.Event {
	// inject source into context for usage in reconciliation logic
	ctx = v1alpha1.WithSource(ctx, src)

	return r.base.ReconcileSource(ctx, r)
}
