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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgotesting "k8s.io/client-go/testing"

	"knative.dev/eventing/pkg/apis/eventing"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	rt "knative.dev/pkg/reconciler/testing"

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
func TestReconcile(t *testing.T, ctor Ctor, src v1alpha1.AWSEventSource,
	adapterFn common.AdapterDeploymentBuilderFunc) {

	assertPopulatedSource(t, src)

	newEventSource := eventSourceCtor(src)
	newAdapterDeployment := adapterDeploymentCtor(adapterFn, src)

	n := newAdapterDeployment().Name

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
				newAdapterDeployment(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, notDeployed),
			}},
			WantEvents: []string{
				createAdapterEvent(n),
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
				newEventSource(withSink, notDeployed),
				newAdapterDeployment(ready),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, deployed),
			}},
		},
		{
			Name: "Adapter becomes NotReady",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, deployed),
				newAdapterDeployment(notReady),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withSink, notDeployed),
			}},
		},
		{
			Name: "Adapter is outdated",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink, deployed),
				newAdapterDeployment(ready, bumpImage),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapterDeployment(ready),
			}},
			WantEvents: []string{
				updateAdapterEvent(n),
			},
		},

		// Errors

		{
			Name: "Sink goes missing",
			Key:  tKey,
			Objects: []runtime.Object{
				/* sink omitted */
				newEventSource(withSink, deployed),
				newAdapterDeployment(ready),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(withoutSink, deployed),
			}},
			WantEvents: []string{
				badSinkEvent(),
			},
			WantErr: true,
		},
		{
			Name: "Fail to create adapter deployment",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("create", "deployments"),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(withSink),
			},
			WantCreates: []runtime.Object{
				newAdapterDeployment(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSource(unknownDeployedWithError, withSink),
			}},
			WantEvents: []string{
				failCreateAdapterEvent(n),
			},
			WantErr: true,
		},
		{
			Name: "Fail to update adapter deployment",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("update", "deployments"),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSource(deployed, withSink),
				newAdapterDeployment(ready, bumpImage),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapterDeployment(ready),
			}},
			WantEvents: []string{
				failUpdateAdapterEvent(n),
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
func deployed(src v1alpha1.AWSEventSource) {
	src.GetStatus().PropagateAvailability(newReadyDeployment())
}

// Deployed: False
func notDeployed(src v1alpha1.AWSEventSource) {
	src.GetStatus().PropagateAvailability(newNotReadyDeployment())
}

// Deployed: Unknown with error
func unknownDeployedWithError(src v1alpha1.AWSEventSource) {
	src.GetStatus().PropagateAvailability(nil)
}

// deleted marks the source as deleted.
func deleted(src v1alpha1.AWSEventSource) {
	t := metav1.Unix(0, 0)
	src.SetDeletionTimestamp(&t)
}

/* Adapter deployment */

// adapterCtorWithOptions is a function that returns a Deployment object with
// modifications applied.
type adapterCtorWithOptions func(...deploymentOption) *appsv1.Deployment

// adapterDeploymentCtor creates a copy of the given Deployment and returns a
// function that can apply options to that Deployment.
func adapterDeploymentCtor(adapterFn common.AdapterDeploymentBuilderFunc,
	src v1alpha1.AWSEventSource) adapterCtorWithOptions {

	// error discarded because the ARN format is already validated during tests
	arn, _ := arn.Parse(src.GetARN())

	return func(opts ...deploymentOption) *appsv1.Deployment {
		d := adapterFn(arn, tSinkURI)

		for _, opt := range opts {
			opt(d)
		}

		return d
	}
}

// deploymentOption is a functional option for a Deployment object.
type deploymentOption func(*appsv1.Deployment)

// Ready: True
func ready(d *appsv1.Deployment) {
	d.Status = appsv1.DeploymentStatus{
		Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		}},
	}
}

func newReadyDeployment() *appsv1.Deployment {
	d := &appsv1.Deployment{}
	ready(d)
	return d
}

// Ready: False
func notReady(d *appsv1.Deployment) {
	d.Status = appsv1.DeploymentStatus{
		Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionFalse,
		}},
	}
}

func newNotReadyDeployment() *appsv1.Deployment {
	d := &appsv1.Deployment{}
	notReady(d)
	return d
}

// bumpImage adds a static suffix to the Deployment's image.
func bumpImage(d *appsv1.Deployment) {
	d.Spec.Template.Spec.Containers[0].Image += "-test"
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

func createAdapterEvent(name string) string {
	return eventtesting.Eventf(corev1.EventTypeNormal, common.ReasonAdapterCreate, "Created adapter Deployment %q", name)
}
func updateAdapterEvent(name string) string {
	return eventtesting.Eventf(corev1.EventTypeNormal, common.ReasonAdapterUpdate, "Updated adapter Deployment %q", name)
}
func failCreateAdapterEvent(name string) string {
	return eventtesting.Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterCreate, "Failed to create adapter Deployment %q: "+
		"inducing failure for create deployments", name)
}
func failUpdateAdapterEvent(name string) string {
	return eventtesting.Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterUpdate, "Failed to update adapter Deployment %q: "+
		"inducing failure for update deployments", name)
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
