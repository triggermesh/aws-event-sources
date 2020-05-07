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
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"

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
	fakek8sinjectionclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awsdynamodbsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	. "github.com/triggermesh/aws-event-sources/pkg/reconciler/testing"
)

const (
	tNs   = "testns"
	tName = "test"
	tKey  = tNs + "/" + tName
	tUID  = types.UID("00000000-0000-0000-0000-000000000000")

	tImg = "registry/image:tag"

	tMetricsDomain = "testing"
)

var tARN = NewARN(dynamodb.ServiceName, "table/triggermeshtest")

var tSinkURI = &apis.URL{
	Scheme: "http",
	Host:   "default.default.svc.example.com",
	Path:   "/",
}

var (
	tAccessKeyIDSelector = &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: "test-secret",
		},
		Key: "keyId",
	}
	tSecretAccessKeySelector = &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: "test-secret",
		},
		Key: "secret",
	}
)

// Test the Reconcile method of the controller.Reconciler implemented by our controller.
//
// The environment for each test case is set up as follows:
//  1. MakeFactory initializes fake clients with the objects declared in the test case
//  2. MakeFactory injects those clients into a context along with fake event recorders, etc.
//  3. A Reconciler is constructed via a Ctor function using the values injected above
//  4. The Reconciler returned by MakeFactory is used to run the test case
func TestReconcile(t *testing.T) {
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
				Object: newEventSourceNotDeployedWithSink(),
			}},
			WantEvents: []string{
				createAdapterEvent(),
			},
		},
		{
			Name: "Source object deletion",
			Key:  tKey,
			Objects: []runtime.Object{
				newDeletedEventSource(),
			},
		},

		// Lifecycle

		{
			Name: "Adapter becomes Ready",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSourceNotDeployedWithSink(),
				newAdapterDeploymentReady(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSourceDeployedWithSink(),
			}},
		},
		{
			Name: "Adapter becomes NotReady",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSourceDeployedWithSink(),
				newAdapterDeploymentNotReady(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSourceNotDeployedWithSink(),
			}},
		},
		{
			Name: "Adapter is outdated",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				newEventSourceDeployedWithSink(),
				setAdapterImage(
					newAdapterDeploymentReady(),
					tImg+":old",
				),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapterDeploymentReady(),
			}},
			WantEvents: []string{
				updateAdapterEvent(),
			},
		},
		{
			Name: "Sink goes missing",
			Key:  tKey,
			Objects: []runtime.Object{
				newEventSourceDeployedWithSink(),
				newAdapterDeploymentReady(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSourceDeployedWithoutSink(),
			}},
			WantEvents: []string{
				badSinkEvent(),
			},
		},

		// Errors

		{
			Name: "Fail to create adapter deployment",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("create", "deployments"),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSourceWithSink(),
			},
			WantCreates: []runtime.Object{
				newAdapterDeployment(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newEventSourceUnknownDeployedWithSink(),
			}},
			WantEvents: []string{
				failCreateAdapterEvent(),
			},
		},
		{
			Name: "Fail to update adapter deployment",
			Key:  tKey,
			WithReactors: []clientgotesting.ReactionFunc{
				rt.InduceFailure("update", "deployments"),
			},
			Objects: []runtime.Object{
				newAdressable(),
				newEventSourceDeployedWithSink(),
				setAdapterImage(
					newAdapterDeploymentReady(),
					tImg+":old",
				),
			},
			WantUpdates: []clientgotesting.UpdateActionImpl{{
				Object: newAdapterDeploymentReady(),
			}},
			WantEvents: []string{
				failUpdateAdapterEvent(),
			},
		},

		// Edge cases

		{
			Name:    "Reconcile a non-existing object",
			Key:     tKey,
			Objects: nil,
			WantErr: false,
		},
	}

	testCases.Test(t, MakeFactory(reconcilerCtor))
}

// reconcilerCtor returns a Ctor for a AWSDynamoDBSource Reconciler.
var reconcilerCtor Ctor = func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
	logger := logging.FromContext(ctx)

	adapterCfg := &adapterConfig{
		Image: tImg,
		// LoggingCfg and MetricsCfg get auto-populated by cmw.Watch
	}

	r := &Reconciler{
		logger:           logger,
		sinkResolver:     resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
		adapterCfg:       adapterCfg,
		deploymentLister: ls.GetDeploymentLister().Deployments,
		deploymentClient: fakek8sinjectionclient.Get(ctx).AppsV1().Deployments,
	}

	// updateAdapterMetricsConfig panics when METRICS_DOMAIN is unset
	defer SetEnvVar(t, metrics.DomainEnv, tMetricsDomain)()

	metricsData := map[string]string{
		"metrics.backend": "prometheus",
	}
	loggingData := map[string]string{
		"zap-logger-config": `{"level": "info"}`,
	}
	cmw := configmap.NewStaticWatcher(
		NewConfigMap(logging.ConfigMapName(), loggingData),
		NewConfigMap(metrics.ConfigMapName(), metricsData),
	)

	cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)
	cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)

	return reconcilerv1alpha1.NewReconciler(ctx, logger,
		fakeinjectionclient.Get(ctx), ls.GetAWSDynamoDBSourceLister(),
		controller.GetEventRecorder(ctx), r)
}

/* Event sources */

// dummy that can be passed to target constructors to indicate that the object
// is to be returned without CloudEvents attributes
const noCEAttributes = iota

// newEventSource returns a test source object with pre-filled attributes.
func newEventSource(skipCEAtrributes ...interface{}) *v1alpha1.AWSDynamoDBSource {
	o := &v1alpha1.AWSDynamoDBSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			UID:       tUID,
		},
		Spec: v1alpha1.AWSDynamoDBSourceSpec{
			ARN: tARN.String(),
			Credentials: v1alpha1.AWSSecurityCredentials{
				AccessKeyID: v1alpha1.ValueFromField{
					ValueFromSecret: tAccessKeyIDSelector,
				},
				SecretAccessKey: v1alpha1.ValueFromField{
					ValueFromSecret: tSecretAccessKeySelector,
				},
			},
		},
	}

	addrGVK := newAdressable().GetGroupVersionKind()
	o.Spec.Sink = duckv1.Destination{
		Ref: &duckv1.KReference{
			APIVersion: addrGVK.GroupVersion().String(),
			Kind:       addrGVK.Kind,
			Name:       tName,
		},
	}

	if len(skipCEAtrributes) == 0 {
		o.Status.CloudEventAttributes = createCloudEventAttributes(tARN)
	}

	o.Status.InitializeConditions()

	return o
}

// Sink: True, Deployed: Unknown
func newEventSourceWithSink() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()
	o.Status.MarkSink(tSinkURI)
	return o
}

// Sink: True, Deployed: True
func newEventSourceDeployedWithSink() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()
	o.Status.PropagateAvailability(newAdapterDeploymentReady())
	o.Status.MarkSink(tSinkURI)
	return o
}

// Sink: False, Deployed: True
func newEventSourceDeployedWithoutSink() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()
	o.Status.MarkNoSink()
	o.Status.PropagateAvailability(newAdapterDeploymentReady())
	return o
}

// Sink: True, Deployed: False
func newEventSourceNotDeployedWithSink() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()
	o.Status.PropagateAvailability(newAdapterDeploymentNotReady())
	o.Status.MarkSink(tSinkURI)
	return o
}

// Sink: True, Deployed: Unknown with error
func newEventSourceUnknownDeployedWithSink() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()
	o.Status.MarkSink(tSinkURI)
	o.Status.PropagateAvailability(nil)
	return o
}

// newDeletedEventSource returns a test source object marked for deletion.
func newDeletedEventSource() *v1alpha1.AWSDynamoDBSource {
	o := newEventSource()

	t := metav1.Unix(0, 0)
	o.SetDeletionTimestamp(&t)

	return o
}

/* Adapter deployment */

// newAdapterDeployment returns a test Deployment object with pre-filled attributes.
func newAdapterDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      adapterName + "-" + tName,
			Labels: map[string]string{
				common.AppNameLabel:      adapterName,
				common.AppInstanceLabel:  tName,
				common.AppComponentLabel: common.AdapterComponent,
				common.AppPartOfLabel:    common.PartOf,
				common.AppManagedByLabel: common.ManagedBy,
			},
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(NewOwnerRefable(
					tName,
					(&v1alpha1.AWSDynamoDBSource{}).GetGroupVersionKind(),
					tUID,
				)),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					common.AppNameLabel:     adapterName,
					common.AppInstanceLabel: tName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						common.AppNameLabel:      adapterName,
						common.AppInstanceLabel:  tName,
						common.AppComponentLabel: common.AdapterComponent,
						common.AppPartOfLabel:    common.PartOf,
						common.AppManagedByLabel: common.ManagedBy,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "adapter", // defaulted by resource package
						Image: tImg,
						Env: []corev1.EnvVar{
							{
								Name:  common.EnvName,
								Value: tName,
							}, {
								Name:  common.EnvNamespace,
								Value: tNs,
							}, {
								Name:  common.EnvSink,
								Value: tSinkURI.String(),
							}, {
								Name:  common.EnvLoggingConfig,
								Value: `{"zap-logger-config":"{\"level\": \"info\"}"}`,
							}, {
								Name: common.EnvMetricsConfig,
								Value: `{"Domain":"` + tMetricsDomain + `",` +
									`"Component":"` + adapterName + `",` +
									`"PrometheusPort":0,` +
									`"ConfigMap":{"metrics.backend":"prometheus"}}`,
							}, {
								Name:  common.EnvARN,
								Value: tARN.String(),
							}, {
								Name: common.EnvAccessKeyID,
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: tAccessKeyIDSelector,
								},
							}, {
								Name: common.EnvSecretAccessKey,
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: tSecretAccessKeySelector,
								},
							},
						},
					}},
				},
			},
		},
	}
}

// Ready: True
func newAdapterDeploymentReady() *appsv1.Deployment {
	d := newAdapterDeployment()
	d.Status = appsv1.DeploymentStatus{
		Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		}},
	}
	return d
}

// Ready: False
func newAdapterDeploymentNotReady() *appsv1.Deployment {
	d := newAdapterDeployment()
	d.Status = appsv1.DeploymentStatus{
		Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionFalse,
		}},
	}
	return d
}

func setAdapterImage(o *appsv1.Deployment, img string) *appsv1.Deployment {
	o.Spec.Template.Spec.Containers[0].Image = img
	return o
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

func createAdapterEvent() string {
	return Eventf(corev1.EventTypeNormal, common.ReasonAdapterCreate, "Created adapter Deployment \"%s-%s\"",
		adapterName, tName)
}
func updateAdapterEvent() string {
	return Eventf(corev1.EventTypeNormal, common.ReasonAdapterUpdate, "Updated adapter Deployment \"%s-%s\"",
		adapterName, tName)
}
func failCreateAdapterEvent() string {
	return Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterCreate, "Failed to create adapter Deployment \"%s-%s\": "+
		"inducing failure for create deployments", adapterName, tName)
}
func failUpdateAdapterEvent() string {
	return Eventf(corev1.EventTypeWarning, common.ReasonFailedAdapterUpdate, "Failed to update adapter Deployment \"%s-%s\": "+
		"inducing failure for update deployments", adapterName, tName)
}
func badSinkEvent() string {
	addrGVK := newAdressable().GetGroupVersionKind()

	return Eventf(corev1.EventTypeWarning, common.ReasonBadSinkURI, "Could not resolve sink URI: "+
		"failed to get ref &ObjectReference{Kind:%s,Namespace:%s,Name:%s,UID:,APIVersion:%s,ResourceVersion:,FieldPath:,}: "+
		"%s %q not found",
		addrGVK.Kind, tNs, tName, addrGVK.GroupVersion().String(),
		eventing.BrokersResource, tName)
}
