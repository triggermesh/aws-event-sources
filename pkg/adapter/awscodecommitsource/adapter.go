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

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

var (
	//syncTime       = 10
	lastCommit     string
	pullRequestIDs []*string //nolint:unused
)

const (
	pushEventType = "push"
	prEventType   = "pull_request"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN           string `envconfig:"ARN" required:"true"`
	Branch        string `envconfig:"BRANCH" required:"true"`
	GitEventTypes string `envconfig:"EVENT_TYPES" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	ccClient codecommitiface.CodeCommitAPI
	ceClient cloudevents.Client

	arn       arn.ARN
	branch    string
	gitEvents string
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region).
		WithMaxRetries(5),
	))

	return &adapter{
		logger: logger,

		ccClient: codecommit.New(cfg),
		ceClient: ceClient,

		arn:       arn,
		branch:    env.Branch,
		gitEvents: env.GitEventTypes,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	if strings.Contains(a.gitEvents, pushEventType) {
		a.logger.Info("Push events enabled")

		branchInfo, err := a.ccClient.GetBranch(&codecommit.GetBranchInput{
			RepositoryName: &a.arn.Resource,
			BranchName:     &a.branch,
		})
		if err != nil {
			a.logger.Fatalw("Failed to retrieve branch info", "error", err)
		}

		lastCommit = *branchInfo.Branch.CommitId
	}

	if strings.Contains(a.gitEvents, prEventType) {
		a.logger.Info("Pull Request events enabled")

		// get pull request IDs
		pullRequestsOutput, err := a.ccClient.ListPullRequests(&codecommit.ListPullRequestsInput{
			RepositoryName: &a.arn.Resource,
		})
		if err != nil {
			a.logger.Fatalw("Failed to retrieve list of pull requests", "error", err)
		}

		pullRequestIDs = pullRequestsOutput.PullRequestIds
	}

	if !strings.Contains(a.gitEvents, pushEventType) && !strings.Contains(a.gitEvents, prEventType) {
		a.logger.Fatalf("Failed to identify event types in %q. Valid values: (push,pull_request)", a.gitEvents)
	}

	processedPullRequests, err := a.preparePullRequests()
	if err != nil {
		a.logger.Errorw("Failed to process pull requests", "error", err)
	}

	//range time.Tick(time.Duration(syncTime) * time.Second)
	for {
		if strings.Contains(a.gitEvents, pushEventType) {
			err := a.processCommits()
			if err != nil {
				a.logger.Errorw("Failed to process commits", "error", err)
			}
		}

		if strings.Contains(a.gitEvents, prEventType) {
			pullRequests, err := a.preparePullRequests()
			if err != nil {
				a.logger.Errorw("Failed to process pull requests", "error", err)
			}

			pullRequests = removeOldPRs(processedPullRequests, pullRequests)

			for _, pr := range pullRequests {
				err = a.sendPREvent(pr)
				if err != nil {
					a.logger.Errorw("Failed to send PR event", "error", err)
				}
				processedPullRequests = append(processedPullRequests, pr)
			}
		}
	}
}

func (a *adapter) processCommits() error {
	branchInfo, err := a.ccClient.GetBranch(&codecommit.GetBranchInput{
		BranchName:     &a.branch,
		RepositoryName: &a.arn.Resource,
	})
	if err != nil {
		return fmt.Errorf("failed to get branch info: %w", err)
	}

	commitOutput, err := a.ccClient.GetCommit(&codecommit.GetCommitInput{
		CommitId:       branchInfo.Branch.CommitId,
		RepositoryName: &a.arn.Resource,
	})
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	if *commitOutput.Commit.CommitId == lastCommit {
		return nil
	}

	lastCommit = *commitOutput.Commit.CommitId

	err = a.sendPushEvent(commitOutput.Commit)
	if err != nil {
		return fmt.Errorf("failed to send push event: %w", err)
	}

	return nil
}

func (a *adapter) preparePullRequests() ([]*codecommit.PullRequest, error) {
	pullRequests := []*codecommit.PullRequest{}

	input := codecommit.ListPullRequestsInput{
		RepositoryName: &a.arn.Resource,
	}

	for {
		//Get pull request IDs
		pullRequestsOutput, err := a.ccClient.ListPullRequests(&input)
		if err != nil {
			return pullRequests, fmt.Errorf("failed to list PRs: %w", err)
		}

		prIDs := []*string{}

		prIDs = append(prIDs, pullRequestsOutput.PullRequestIds...)

		for _, id := range prIDs {
			pri := codecommit.GetPullRequestInput{PullRequestId: id}

			prInfo, err := a.ccClient.GetPullRequest(&pri)
			if err != nil {
				return pullRequests, fmt.Errorf("failed to get PR info: %w", err)
			}

			pullRequests = append(pullRequests, prInfo.PullRequest)
		}

		if pullRequestsOutput.NextToken == nil {
			break
		}

		input.NextToken = pullRequestsOutput.NextToken
	}

	return pullRequests, nil
}

// sendPushEvent sends an event containing data about a git commit that was
// pushed to a branch.
func (a *adapter) sendPushEvent(commit *codecommit.Commit) error {
	a.logger.Info("Sending Push event")

	data := &PushEvent{
		Commit:           commit,
		CommitRepository: &a.arn.Resource,
		CommitBranch:     &a.branch,
		EventSource:      aws.String("aws:codecommit"),
		AwsRegion:        &a.arn.Region,
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, pushEventType))
	event.SetSubject(fmt.Sprintf("%s/%s", a.arn.Resource, a.branch))
	event.SetSource(a.arn.String())
	event.SetID(*commit.CommitId)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

func (a *adapter) sendPREvent(pullRequest *codecommit.PullRequest) error {
	a.logger.Info("Sending Pull Request event")

	data := &PullRequestEvent{
		PullRequest: pullRequest,
		Repository:  &a.arn.Resource,
		Branch:      &a.branch,
		EventSource: aws.String("aws:codecommit"),
		AwsRegion:   &a.arn.Region,
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, prEventType))
	event.SetSubject(a.branch)
	event.SetSource(a.arn.String())
	event.SetID(*pullRequest.PullRequestId)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

func removeOldPRs(oldPrs, newPrs []*codecommit.PullRequest) []*codecommit.PullRequest {
	dct := make(map[string]*codecommit.PullRequest)
	for _, oldPR := range oldPrs {
		dct[*oldPR.PullRequestId] = oldPR
	}

	res := make([]*codecommit.PullRequest, 0)

	for _, newPR := range newPrs {
		if v, exist := dct[*newPR.PullRequestId]; !exist {
			res = append(res, newPR)
			continue
		} else {
			if *newPR.PullRequestStatus == *v.PullRequestStatus {
				continue
			}
			res = append(res, newPR)
		}
	}
	return res
}
