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

package common

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	k8sclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformerv1 "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
)

// GenericDeploymentReconciler contains interfaces shared across Deployment reconcilers.
type GenericDeploymentReconciler struct {
	// URI resolver for sinks
	SinkResolver *resolver.URIResolver
	// API clients
	Client func(namespace string) appsclientv1.DeploymentInterface
	// objects listers
	Lister func(namespace string) appslistersv1.DeploymentNamespaceLister
}

// NewGenericDeploymentReconciler creates a new GenericDeploymentReconciler and
// attaches a default event handler to its Deployment informer.
func NewGenericDeploymentReconciler(ctx context.Context, gvk schema.GroupVersionKind,
	resolverCallback func(types.NamespacedName),
	adapterHandlerFn func(obj interface{}),
) GenericDeploymentReconciler {

	informer := deploymentinformerv1.Get(ctx)

	r := GenericDeploymentReconciler{
		SinkResolver: resolver.NewURIResolver(ctx, resolverCallback),
		Client:       k8sclient.Get(ctx).AppsV1().Deployments,
		Lister:       informer.Lister().Deployments,
	}

	informer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK(gvk),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	return r
}
