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

package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awscloudwatchlogssource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awscloudwatchsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awscodecommitsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awscognitoidentitysource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awscognitouserpoolsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awsdynamodbsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awsiotsource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awskinesissource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awsperformanceinsightssource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awss3source"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awssnssource"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/awssqssource"
)

func main() {
	sharedmain.Main("aws-event-sources-controller",
		awscloudwatchlogssource.NewController,
		awscloudwatchsource.NewController,
		awscodecommitsource.NewController,
		awscognitoidentitysource.NewController,
		awscognitouserpoolsource.NewController,
		awsdynamodbsource.NewController,
		awsiotsource.NewController,
		awskinesissource.NewController,
		awsperformanceinsightssource.NewController,
		awss3source.NewController,
		awssnssource.NewController,
		awssqssource.NewController,
	)
}
