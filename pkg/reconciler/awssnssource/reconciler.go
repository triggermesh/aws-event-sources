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

	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awssnssource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
)

// Reconciler implements controller.Reconciler for the event source type.
type Reconciler struct {
	base       common.GenericServiceReconciler
	adapterCfg *adapterConfig

	// API clients
	secretsCli func(namespace string) coreclientv1.SecretInterface
}

// Check that our Reconciler implements Interface
var _ reconcilerv1alpha1.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
/* TODO(antoineco): reconcile SNS subscriptions.
   Ref. https://github.com/triggermesh/aws-event-sources/issues/157
*/
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.AWSSNSSource) reconciler.Event {
	// inject source into context for usage in reconciliation logic
	ctx = v1alpha1.WithSource(ctx, src)

	if err := r.base.ReconcileSource(ctx, adapterServiceBuilder(src, r.adapterCfg)); err != nil {
		return fmt.Errorf("failed to reconcile source: %w", err)
	}

	return r.ensureSubscribed(ctx)
}
