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

package awskinesissource

import (
	"context"

	"github.com/kelseyhightower/envconfig"

	"k8s.io/client-go/tools/cache"

	k8sclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformerv1 "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	informerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/informers/sources/v1alpha1/awskinesissource"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awskinesissource"
)

// NewController creates a Reconciler for the event source and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	// Calling envconfig.Process() with a prefix appends that prefix
	// (uppercased) to the Go field name, e.g. MYSOURCE_IMAGE.
	adapterCfg := &adapterConfig{}
	envconfig.MustProcess(adapterName, adapterCfg)

	logger := logging.FromContext(ctx)

	sourceInformer := informerv1alpha1.Get(ctx)
	deploymentInformer := deploymentinformerv1.Get(ctx)

	r := &Reconciler{
		logger:           logger,
		adapterCfg:       adapterCfg,
		deploymentClient: k8sclient.Get(ctx).AppsV1().Deployments,
		deploymentLister: deploymentInformer.Lister().Deployments,
	}
	impl := reconcilerv1alpha1.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	sourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK((&v1alpha1.AWSKinesisSource{}).GetGroupVersionKind()),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)
	cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)

	return impl
}
