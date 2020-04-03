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

package reconciler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/controller"
)

// Reasons for API Events
const (
	// ReasonCreate is used when an object is successfully created.
	ReasonCreate = "Create"
	// ReasonUpdate is used when an object is successfully updated.
	ReasonUpdate = "Update"
	// ReasonFailedCreate is used when an object creation fails.
	ReasonFailedCreate = "FailedCreate"
	// ReasonFailedUpdate is used when an object update fails.
	ReasonFailedUpdate = "FailedUpdate"

	// ReasonBadSinkURI is used when the URI of a sink can't be determined.
	ReasonBadSinkURI = "BadSinkURI"
)

// Event records a normal event for an API object.
func Event(ctx context.Context, obj runtime.Object, reason, msgFmt string, args ...interface{}) {
	recordEvent(ctx, obj, corev1.EventTypeNormal, reason, msgFmt, args...)
}

// EventWarn records a warning event for an API object.
func EventWarn(ctx context.Context, obj runtime.Object, reason, msgFmt string, args ...interface{}) {
	recordEvent(ctx, obj, corev1.EventTypeWarning, reason, msgFmt, args...)
}

func recordEvent(ctx context.Context, obj runtime.Object, typ, reason, msgFmt string, args ...interface{}) {
	controller.GetEventRecorder(ctx).Eventf(obj, typ, reason, msgFmt, args...)
}
