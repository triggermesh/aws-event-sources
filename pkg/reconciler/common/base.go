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

package common

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacclientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	rbaclistersv1 "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"

	k8sclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformerv1 "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	sainformerv1 "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	rbinformerv1 "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
	servingclientv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	serviceinformerv1 "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"
	servinglistersv1 "knative.dev/serving/pkg/client/listers/serving/v1"
)

// GenericDeploymentReconciler contains interfaces shared across Deployment reconcilers.
type GenericDeploymentReconciler struct {
	// URI resolver for sinks
	SinkResolver *resolver.URIResolver
	// API clients
	Client    func(namespace string) appsclientv1.DeploymentInterface
	PodClient func(namespace string) coreclientv1.PodInterface
	// objects listers
	Lister func(namespace string) appslistersv1.DeploymentNamespaceLister

	*GenericRBACReconciler
}

// GenericServiceReconciler contains interfaces shared across Service reconcilers.
type GenericServiceReconciler struct {
	// URI resolver for sinks
	SinkResolver *resolver.URIResolver
	// API clients
	Client func(namespace string) servingclientv1.ServiceInterface
	// objects listers
	Lister func(namespace string) servinglistersv1.ServiceNamespaceLister

	*GenericRBACReconciler
}

// GenericRBACReconciler reconciles RBAC objects for source adapters.
type GenericRBACReconciler struct {
	// API clients
	SAClient func(namespace string) coreclientv1.ServiceAccountInterface
	RBClient func(namespace string) rbacclientv1.RoleBindingInterface
	// objects listers
	SALister func(namespace string) corelistersv1.ServiceAccountNamespaceLister
	RBLister func(namespace string) rbaclistersv1.RoleBindingNamespaceLister
}

// NewGenericDeploymentReconciler creates a new GenericDeploymentReconciler and
// attaches a default event handler to its Deployment informer.
func NewGenericDeploymentReconciler(ctx context.Context, gvk schema.GroupVersionKind,
	resolverCallback func(types.NamespacedName),
	adapterHandlerFn func(obj interface{}),
) GenericDeploymentReconciler {

	informer := deploymentinformerv1.Get(ctx)

	r := GenericDeploymentReconciler{
		SinkResolver:          resolver.NewURIResolver(ctx, resolverCallback),
		Client:                k8sclient.Get(ctx).AppsV1().Deployments,
		PodClient:             k8sclient.Get(ctx).CoreV1().Pods,
		Lister:                informer.Lister().Deployments,
		GenericRBACReconciler: NewGenericRbacReconciler(ctx),
	}

	informer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK(gvk),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	return r
}

// NewGenericServiceReconciler creates a new GenericServiceReconciler and
// attaches a default event handler to its Service informer.
func NewGenericServiceReconciler(ctx context.Context, gvk schema.GroupVersionKind,
	resolverCallback func(types.NamespacedName),
	adapterHandlerFn func(obj interface{}),
) GenericServiceReconciler {

	informer := serviceinformerv1.Get(ctx)

	r := GenericServiceReconciler{
		SinkResolver:          resolver.NewURIResolver(ctx, resolverCallback),
		Client:                servingclient.Get(ctx).ServingV1().Services,
		Lister:                informer.Lister().Services,
		GenericRBACReconciler: NewGenericRbacReconciler(ctx),
	}

	informer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK(gvk),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	return r
}

// NewGenericRbacReconciler creates a new GenericRbacReconciler.
func NewGenericRbacReconciler(ctx context.Context) *GenericRBACReconciler {
	return &GenericRBACReconciler{
		SAClient: k8sclient.Get(ctx).CoreV1().ServiceAccounts,
		RBClient: k8sclient.Get(ctx).RbacV1().RoleBindings,
		SALister: sainformerv1.Get(ctx).Lister().ServiceAccounts,
		RBLister: rbinformerv1.Get(ctx).Lister().RoleBindings,
	}
}
