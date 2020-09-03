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

// Copied from https://github.com/knative/eventing/blob/v0.17.3/test/lib/duck/resource_checks.go and modified to avoid
// dependencies on knative.dev/pkg/test, which redefines the '-kubeconfig' flag.

package sources

import (
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	interval = 1 * time.Second
	timeout  = 2 * time.Minute
)

// WaitForResourceReady polls the status of the MetaResource from client
// every interval until isResourceReady returns `true` indicating
// it is done, returns an error or timeout.
func WaitForResourceReady(dynamicClient dynamic.Interface, obj *unstructured.Unstructured) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		untyped, err := GetGenericObject(dynamicClient, obj, &duckv1beta1.KResource{})
		return isResourceReady(untyped, err)
	})
}

// isResourceReady leverage duck-type to check if the given MetaResource is in ready state
func isResourceReady(obj runtime.Object, err error) (bool, error) {
	if k8serrors.IsNotFound(err) {
		// Return false as we are not done yet.
		// We swallow the error to keep on polling.
		// It should only happen if we wait for the auto-created resources, like default Broker.
		return false, nil
	} else if err != nil {
		// Return error to stop the polling.
		return false, err
	}

	kr := obj.(*duckv1beta1.KResource)
	ready := kr.Status.GetCondition(apis.ConditionReady)
	return ready != nil && ready.IsTrue(), nil
}

// GetGenericObject returns a generic object representing a Kubernetes resource.
// Callers can cast this returned object to other objects that implement the corresponding duck-type.
func GetGenericObject(
	dynamicClient dynamic.Interface,
	obj *unstructured.Unstructured,
	rtype apis.Listable,
) (runtime.Object, error) {
	// get the resource's namespace and gvr
	gvr, _ := meta.UnsafeGuessKindToResource(obj.GroupVersionKind())
	u, err := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(obj.GetName(), metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	res := rtype.DeepCopyObject()
	if err := duck.FromUnstructured(u, res); err != nil {
		return nil, err
	}

	return res, nil
}
