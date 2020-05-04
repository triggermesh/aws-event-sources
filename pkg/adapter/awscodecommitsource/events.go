/*
Copyright (c) 2019-2020 TriggerMesh Inc.

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

package awscodecommitsource

import "github.com/aws/aws-sdk-go/service/codecommit"

// PushEvent represent a push event from CodeCommit source.
type PushEvent struct {
	Commit           *codecommit.Commit `json:"commit"`
	CommitRepository *string            `json:"commitRepository"`
	CommitBranch     *string            `json:"commitBranch"`
	CommitHash       *string            `json:"commitHash"`
	EventSource      *string            `json:"eventSource"`
	AwsRegion        *string            `json:"awsRegion"`
}

// PullRequestEvent represent a PR event from CodeCommit source.
type PullRequestEvent struct {
	PullRequest *codecommit.PullRequest `json:"pullRequest"`
	EventType   *string                 `json:"eventType"`
	Repository  *string                 `json:"repository"`
	Branch      *string                 `json:"branch"`
	EventSource *string                 `json:"eventSource"`
	AwsRegion   *string                 `json:"awsRegion"`
}
