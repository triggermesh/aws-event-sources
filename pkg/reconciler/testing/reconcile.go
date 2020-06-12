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

package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgotesting "k8s.io/client-go/testing"

	"knative.dev/eventing/pkg/apis/eventing"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	rt "knative.dev/pkg/reconciler/testing"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	eventtesting "github.com/triggermesh/aws-event-sources/pkg/reconciler/common/event/testing"
)

const (
	tNs   = "testns"
	tName = "test"
	tKey  = tNs + "/" + tName
	tUID  = types.UID("00000000-0000-0000-0000-000000000000")
)

var tSinkURI = &apis.URL{
	Scheme: "http",
	Host:   "default.default.svc.example.com",
	Path:   "/",
}

// Test the Reconcile method of the controller.Reconciler implemented by controllers.
//
// The environment for each test case is set up as follows:
//  1. MakeFactory initializes fake clients with the objects declared in the test case
//  2. MakeFactory injects those clients into a context along with fake event recorders, etc.
//  3. A Reconciler is constructed via a Ctor function using the values injected above
//  4. The Reconciler returned by MakeFactory is used to run the test case
func TestReconcile(t *testing.T, ctor Ctor, src v1alpha1.AWSEventSource, adapterFn interface{}) {
	assertPopulatedSource(t, src)

	newEventSource := eventSourceCtor(src)
	newAdapter := adapterCtor(adapterFn, src)

	a := newAdapter()
	n, k, r := nameKindAndResource(a)

	testCases := rt.TableTest{
		// Creation/Deletion

		{
			Name: "Source object creation",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(noCEAttributes),
			},
			WantCreates: []runtime.Object{
				newAdapter(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, notDeployed(a)),
			}},
			WantEvents: []string{
				createAdapterEvent(n, k),
			},
		},
		{
			Name: "Source object deletion",
			Key:  tKey,
			Objects: []runtime.Object{
				newEventSource(deleted),
			},
		},

		// Lifecycle

		{
			Name: "Adapter becomes Ready",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, notDeployed(a)),
				newAdapter(ready),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, deployed(a)),
			}},
		},
		{
			Name: "Adapter becomes NotReady",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, deployed(a)),
				newAdapter(notReady),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, notDeployed(a)),
			}},
		},
		{
			Name: "Adapter is outdated",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, deployed(a)),
				newAdapter(ready, bumpImage),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapter(ready),
			}},
			WantEvents: []string{
				updateAdapterEvent(n, k),
			},
		},

		// Errors

		{
			Name: "Sink goes missing",
			Key:  tKey,
			Objects: []runtime.Object{
				/* sink omitted */
				newEventSource(withSink, deployed(a)),
				newAdapter(ready),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withoutSink, deployed(a)),
			}},
			WantEvents: []string{
				badSinkEvent(),
			},
			WantErr: true,
		},
		{
			Name: "Fail to create adapter",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("create", r),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink),
			},
			WantCreates: []runtime.Object{
				newAdapter(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(unknownDeployedWithError(a), withSink),
			}},
			WantEvents: []string{
				failCreateAdapterEvent(n, k, r),
			},
			WantErr: true,
		},
		{
			Name: "Fail to update adapter",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("update", r),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, deployed(a)),
				newAdapter(ready, bumpImage),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapter(ready),
			}},
			WantEvents: []string{
				failUpdateAdapterEvent(n, k, r),
			},
			WantErr: true,
		},

		// Edge cases

		{
			Name:    "Reconcile a non-existing object",
			Key:     tKey,
			Objects: nil,
			WantErr: false,
		},
	}

	testCases.Test(t, MakeFactory(ctor))
}

// assertPopulatedSource asserts that all source attributes required in
// reconciliation tests are populated and valid.
func assertPopulatedSource(t *testing.T, src v1alpha1.AWSEventSource) {
	t.Helper()

	_, err := arn.Parse(src.GetARN())
	assert.NoError(t, err, "Provided source object should have a valid ARN")

	// used to generate the adapter's owner reference
	assert.NotEmpty(t, src.GetNamespace())
	assert.NotEmpty(t, src.GetName())
	assert.NotEmpty(t, src.GetUID())

	assert.NotEmpty(t, src.GetSink().Ref, "Provided source should reference a sink")
	assert.NotEmpty(t, src.GetEventTypes(), "Provided source should declare its event types")
	assert.NotEmpty(t, src.GetStatus().Status.Conditions, "Provided source should have initialized conditions")
}

func nameKindAndResource(object runtime.Object) (string /*name*/, string /*kind*/, string /*resource*/) {
	metaObj, _ := meta.Accessor(object)
	name := metaObj.GetName()

	var kind, resource string

	switch object.(type) {
	case *appsv1.Deployment:
		kind = "Deployment"
		resource = "deployments"
	case *servingv1.Service:
		kind = "Service"
		resource = "services"
	}

	return name, kind, resource
}

/* Event sources */

// Populate populates an event source with generic attributes.
func Populate(srcCpy v1alpha1.AWSEventSource) {
	srcCpy.SetNamespace(tNs)
	srcCpy.SetName(tName)
	srcCpy.SetUID(tUID)

	addr := newAdressable()
	addrGVK := addr.GetGroupVersionKind()

	srcCpy.GetSink().Ref = &duckv1.KReference{
		APIVersion: addrGVK.GroupVersion().String(),
		Kind:       addrGVK.Kind,
		Name:       addr.GetName(),
	}

	// error discarded because the ARN format is already validated during tests
	arn, _ := arn.Parse(srcCpy.GetARN())

	status := srcCpy.GetStatus()
	status.CloudEventAttributes = common.CreateCloudEventAttributes(arn, srcCpy.GetEventTypes())
	status.InitializeConditions()
}

// sourceCtorWithOptions is a function that returns a source object with
// modifications applied.
type sourceCtorWithOptions func(...sourceOption) v1alpha1.AWSEventSource

// eventSourceCtor creates a copy of the given source object and returns a
// function that can be invoked to return that source, with the possibility to
// apply options to it.
func eventSourceCtor(src v1alpha1.AWSEventSource) sourceCtorWithOptions {
	return func(opts ...sourceOption) v1alpha1.AWSEventSource {
		srcCpy := src.DeepCopyObject().(v1alpha1.AWSEventSource)

		for _, opt := range opts {
			opt(srcCpy)
		}

		return srcCpy
	}
}

// sourceOption is a functional option for a source interface.
type sourceOption func(v1alpha1.AWSEventSource)

// noCEAttributes sets empty CE attributes. Simulates the creation of a new source.
func noCEAttributes(src v1alpha1.AWSEventSource) {
	src.GetStatus().CloudEventAttributes = nil
}

// Sink: True
func withSink(src v1alpha1.AWSEventSource) {
	src.GetStatus().MarkSink(tSinkURI)
}

// Sink: False
func withoutSink(src v1alpha1.AWSEventSource) {
	src.GetStatus().MarkNoSink()
}

// Deployed: True
func deployed(adapter runtime.Object) sourceOption {
	adapter = adapter.DeepCopyObject()
	ready(adapter)

	return func(src v1alpha1.AWSEventSource) {
		src.GetStatus().PropagateAvailability(adapter)
	}
}

// Deployed: False
func notDeployed(adapter runtime.Object) sourceOption {
	adapter = adapter.DeepCopyObject()
	notReady(adapter)

	return func(src v1alpha1.AWSEventSource) {
		src.GetStatus().PropagateAvailability(adapter)
	}
}

// Deployed: Unknown with error
func unknownDeployedWithError(adapter runtime.Object) sourceOption {
	var nilObj runtime.Object

	switch adapter.(type) {
	case *appsv1.Deployment:
		nilObj = (*appsv1.Deployment)(nil)
	case *servingv1.Service:
		nilObj = (*servingv1.Service)(nil)
	}

	return func(src v1alpha1.AWSEventSource) {
		src.GetStatus().PropagateAvailability(nilObj)
	}
}

// deleted marks the source as deleted.
func deleted(src v1alpha1.AWSEventSource) {
	t := metav1.Unix(0, 0)
	src.SetDeletionTimestamp(&t)
}

/* Adapter */

// adapterCtorWithOptions is a function that returns a runtime object with
// modifications applied.
type adapterCtorWithOptions func(...adapterOption) runtime.Object

// adapterCtor creates a copy of the given adapter object and returns a
// function that can apply options to that object.
func adapterCtor(adapterFn interface{}, src v1alpha1.AWSEventSource) adapterCtorWithOptions {
	// error discarded because the ARN format is already validated during tests
	arn, _ := arn.Parse(src.GetARN())

	return func(opts ...adapterOption) runtime.Object {
		var obj runtime.Object

		switch typedAdapterFn := adapterFn.(type) {
		case common.AdapterDeploymentBuilderFunc:
			obj = typedAdapterFn(arn, tSinkURI)
		case common.AdapterServiceBuilderFunc:
			obj = typedAdapterFn(arn, tSinkURI)
		}

		for _, opt := range opts {
			opt(obj)
		}

		return obj
	}
}

// adapterOption is a functional option for an adapter object.
type adapterOption func(runtime.Object)

// Ready: True
func ready(object runtime.Object) {
	switch o := object.(type) {
	case *appsv1.Deployment:
		o.Status = appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionTrue,
			}},
		}
	case *servingv1.Service:
		o.Status.SetConditions(apis.Conditions{{
			Type:   v1alpha1.ConditionReady,
			Status: corev1.ConditionTrue,
		}})
	}
}

// Ready: False
func notReady(object runtime.Object) {
	switch o := object.(type) {
	case *appsv1.Deployment:
		o.Status = appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionFalse,
			}},
		}
	case *servingv1.Service:
		o.Status.SetConditions(apis.Conditions{{
			Type:   v1alpha1.ConditionReady,
			Status: corev1.ConditionFalse,
		}})
	}
}

// bumpImage adds a static suffix to the Deployment's image.
func bumpImage(object runtime.Object) {
	switch o := object.(type) {
	case *appsv1.Deployment:
		o.Spec.Template.Spec.Containers[0].Image += "-test"
	case *servingv1.Service:
		o.Spec.Template.Spec.Containers[0].Image += "-test"
	}
}

/* Event sink */

// newAdressable returns a test Addressable to be used as a sink.
func newAdressable() *eventingv1beta1.Broker {
	return &eventingv1beta1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Status: eventingv1beta1.BrokerStatus{
			Address: duckv1.Addressable{
				URL: tSinkURI,
			},
		},
	}
}

/* Events */

func createAdapterEvent(name, kind string) string {
	return eventtesting.Eventf(corev1.EventTypeNormal, common.ReasonAdapterCreate, "Created adapter %s %q", kind, name)
}
func updateAdapterEvent(name, kind string) string {
	return eventtesting.Eventf(corev1.EventTypeNormal, common.ReasonAdapterUpdate, "Updated adapter %s %q", kind, name)
}
func failCreateAdapterEvent(name, kind, resource string) string {
	return eventtesting.Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterCreate, "Failed to create adapter %s %q: "+
		"inducing failure for create %s", kind, name, resource)
}
func failUpdateAdapterEvent(name, kind, resource string) string {
	return eventtesting.Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterUpdate, "Failed to update adapter %s %q: "+
		"inducing failure for update %s", kind, name, resource)
}
func badSinkEvent() string {
	addrGVK := newAdressable().GetGroupVersionKind()

	// FIXME: the event reason is "InternalError" instead of the expected common.ReasonBadSinkURI
	// because controller.NewPermanentError does not use Go's error wrapping.
	return eventtesting.Eventf(corev1.EventTypeWarning, "InternalError", "Could not resolve sink URI: "+
		"failed to get ref &ObjectReference{Kind:%s,Namespace:%s,Name:%s,UID:,APIVersion:%s,ResourceVersion:,FieldPath:,}: "+
		"%s %q not found",
		addrGVK.Kind, tNs, tName, addrGVK.GroupVersion().String(),
		eventing.BrokersResource, tName)
}
