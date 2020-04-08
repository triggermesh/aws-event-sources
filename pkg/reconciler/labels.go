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

// Kubernetes recommended labels
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
const (
	// AppNameLabel is the name of the application.
	AppNameLabel = "app.kubernetes.io/name"
	// AppInstanceLabel is a unique name identifying the instance of an application.
	AppInstanceLabel = "app.kubernetes.io/instance"
	// AppVersionLabel is the current version of the application.
	AppVersionLabel = "app.kubernetes.io/version"
	// AppComponentLabel is the component within the architecture.
	AppComponentLabel = "app.kubernetes.io/component"
	// AppPartOfLabel is the name of a higher level application this one is part of.
	AppPartOfLabel = "app.kubernetes.io/part-of"
	// AppManagedByLabel is the tool being used to manage the operation of an application.
	AppManagedByLabel = "app.kubernetes.io/managed-by"
)

// Common label values
const (
	SourcesGroup     = "aws-event-sources"
	ManagedBy        = "aws-sources-controller"
	AdapterComponent = "adapter"
)
