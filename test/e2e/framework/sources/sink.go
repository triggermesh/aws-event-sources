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

package sources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/deployment"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/ptr"
)

const (
	eventDisplayName           = "event-display"
	eventDisplayContainerImage = "gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display:latest"
)

// CreateEventDisplaySink creates an event-display event sink.
func CreateEventDisplaySink(cli kubernetes.Interface, namespace string) duckv1.Destination {
	svc, depl := eventDisplayServiceAndDeployment(namespace)

	if _, err := cli.CoreV1().Services(namespace).Create(svc); err != nil {
		framework.Failf("Failed to create event-display Service: %s", err)
	}

	if _, err := cli.AppsV1().Deployments(namespace).Create(depl); err != nil {
		framework.Failf("Failed to create event-display Deployment: %s", err)
	}

	if err := deployment.WaitForDeploymentComplete(cli, depl); err != nil {
		framework.Failf("Error waiting for event-display Deployment to become ready: %s", err)
	}

	return duckv1.Destination{
		Ref: &duckv1.KReference{
			APIVersion: "v1",
			Kind:       "Service",
			Name:       svc.Name,
		},
	}
}

// eventDisplayServiceAndDeployment returns a Service object and a Deployment
// object for the event-display event sink.
func eventDisplayServiceAndDeployment(namespace string) (*corev1.Service, *appsv1.Deployment) {
	const portName = "http"

	lbls := labels.Set{
		"app.kubernetes.io/name":       eventDisplayName,
		"app.kubernetes.io/managed-by": "e2e-testing",
	}

	metadata := metav1.ObjectMeta{
		Namespace: namespace,
		Name:      eventDisplayName,
		Labels:    lbls,
	}

	return &corev1.Service{
			ObjectMeta: metadata,
			Spec: corev1.ServiceSpec{
				Selector: lbls,
				Ports: []corev1.ServicePort{{
					Name:       portName,
					Port:       80,
					TargetPort: intstr.FromString(portName),
				}},
			},
		},

		&appsv1.Deployment{
			ObjectMeta: metadata,
			Spec: appsv1.DeploymentSpec{
				// must not be nil to avoid crashing during WaitForDeploymentComplete
				Replicas: ptr.Int32(1),
				Selector: metav1.SetAsLabelSelector(lbls),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: lbls,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  eventDisplayName,
							Image: eventDisplayContainerImage,
							Ports: []corev1.ContainerPort{{
								Name:          portName,
								ContainerPort: 8080,
							}},
						}},
					},
				},
			},
		}
}
