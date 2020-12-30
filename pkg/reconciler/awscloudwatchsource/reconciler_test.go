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

package awscloudwatchsource

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"knative.dev/eventing/pkg/reconciler/source"
	fakek8sinjectionclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/aws-event-sources/pkg/apis"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/awscloudwatchsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common"
	. "github.com/triggermesh/aws-event-sources/pkg/reconciler/testing"
)

func TestReconcileSource(t *testing.T) {
	adapterCfg := &adapterConfig{
		Image:   "registry/image:tag",
		configs: &source.EmptyVarsGenerator{},
	}

	var (
		ctor      = reconcilerCtor(adapterCfg)
		src       = newEventSource()
		adapterFn = adapterDeploymentBuilder(src, adapterCfg)
	)

	TestReconcile(t, ctor, src, adapterFn)
}

// reconcilerCtor returns a Ctor for a AWSCloudWatchSource Reconciler.
func reconcilerCtor(cfg *adapterConfig) Ctor {
	return func(t *testing.T, ctx context.Context, _ *rt.TableRow, ls *Listers) controller.Reconciler {
		base := common.GenericDeploymentReconciler{
			SinkResolver: resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			Lister:       ls.GetDeploymentLister().Deployments,
			Client:       fakek8sinjectionclient.Get(ctx).AppsV1().Deployments,
			PodClient:    fakek8sinjectionclient.Get(ctx).CoreV1().Pods,
		}

		r := &Reconciler{
			base:       base,
			adapterCfg: cfg,
		}

		return reconcilerv1alpha1.NewReconciler(ctx, logging.FromContext(ctx),
			fakeinjectionclient.Get(ctx), ls.GetAWSCloudWatchSourceLister(),
			controller.GetEventRecorder(ctx), r)
	}
}

// newEventSource returns a populated source object.
func newEventSource() *v1alpha1.AWSCloudWatchSource {
	pollingInterval := apis.Duration(5 * time.Minute)

	src := &v1alpha1.AWSCloudWatchSource{
		Spec: v1alpha1.AWSCloudWatchSourceSpec{
			Region: "us-west-2",
			MetricQueries: []v1alpha1.AWSCloudWatchMetricQuery{{
				Name:       "testquery",
				Expression: nil,
				Metric: &v1alpha1.AWSCloudWatchMetricStat{
					Metric: v1alpha1.AWSCloudWatchMetric{
						Dimensions: []v1alpha1.AWSCloudWatchMetricDimension{{
							Name:  "FunctionName",
							Value: "makemoney",
						}},
						MetricName: "Duration",
						Namespace:  "AWS/Lambda",
					},
					Period: 60,
					Stat:   "sum",
					Unit:   "seconds",
				},
			}},
			PollingInterval: &pollingInterval,
			Credentials: v1alpha1.AWSSecurityCredentials{
				AccessKeyID: v1alpha1.ValueFromField{
					ValueFromSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "keyId",
					},
				},
				SecretAccessKey: v1alpha1.ValueFromField{
					ValueFromSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "secret",
					},
				},
			},
		},
		Status: v1alpha1.EventSourceStatus{},
	}

	Populate(src)

	return src
}
