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

package awscodecommitsource

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

var (
	//syncTime       = 10
	lastCommit     string
	pullRequestIDs []*string
)

const (
	pushEventType = "push"
	prEventType   = "pull_request"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	Repo                   string `envconfig:"REPO" required:"true"`
	RepoBranch             string `envconfig:"BRANCH" required:"true"`
	GitEvents              string `envconfig:"EVENTS" required:"true"`
	AWSRegion              string `envconfig:"AWS_REGION" required:"true"`
	AccountAccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AccountSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	ccClient codecommitiface.CodeCommitAPI
	ceClient cloudevents.Client

	repo                   string
	repoBranch             string
	gitEvents              string
	awsRegion              string
	accountAccessKeyID     string
	accountSecretAccessKey string
}

func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor,
	ceClient cloudevents.Client) pkgadapter.Adapter {

	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	// create CodeCommit client
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(env.AWSRegion),
		Credentials: credentials.NewStaticCredentials(env.AccountAccessKeyID, env.AccountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logger.Fatalw("Failed to create CodeCommit client", "error", err)
	}

	return &adapter{
		logger: logger,

		ccClient: codecommit.New(sess),
		ceClient: ceClient,

		repo:                   env.Repo,
		repoBranch:             env.RepoBranch,
		gitEvents:              env.GitEvents,
		awsRegion:              env.AWSRegion,
		accountAccessKeyID:     env.AccountAccessKeyID,
		accountSecretAccessKey: env.AccountSecretAccessKey,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	if strings.Contains(a.gitEvents, pushEventType) {
		a.logger.Info("Push events enabled")

		branchInfo, err := a.ccClient.GetBranch(&codecommit.GetBranchInput{
			RepositoryName: aws.String(a.repo),
			BranchName:     aws.String(a.repoBranch),
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
			RepositoryName: aws.String(a.repo),
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
		BranchName:     aws.String(a.repoBranch),
		RepositoryName: aws.String(a.repo),
	})
	if err != nil {
		return fmt.Errorf("failed to get branch info: %w", err)
	}

	commitOutput, err := a.ccClient.GetCommit(&codecommit.GetCommitInput{
		CommitId:       branchInfo.Branch.CommitId,
		RepositoryName: aws.String(a.repo),
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
		RepositoryName: aws.String(a.repo),
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
// pushed to a branch
func (a *adapter) sendPushEvent(commit *codecommit.Commit) error {
	a.logger.Info("Sending Push event")

	data := &PushEvent{
		Commit:           commit,
		CommitRepository: aws.String(a.repo),
		CommitBranch:     aws.String(a.repoBranch),
		EventSource:      aws.String("aws:codecommit"),
		AwsRegion:        aws.String(a.awsRegion),
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSCodeCommitEventType(pushEventType))
	event.SetSubject(fmt.Sprintf("%s/%s", a.repo, a.repoBranch))
	event.SetSource(v1alpha1.AWSCodeCommitEventSource(a.awsRegion, a.repo))
	event.SetID(*commit.CommitId)
	event.SetData(cloudevents.ApplicationJSON, data)

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

func (a *adapter) sendPREvent(pullRequest *codecommit.PullRequest) error {
	a.logger.Info("Sending Pull Request event")

	data := &PullRequestEvent{
		PullRequest: pullRequest,
		Repository:  aws.String(a.repo),
		Branch:      aws.String(a.repoBranch),
		EventSource: aws.String("aws:codecommit"),
		AwsRegion:   aws.String(a.awsRegion),
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSCodeCommitEventType(prEventType))
	event.SetSubject(fmt.Sprintf("%s/%s", a.repo, a.repoBranch))
	event.SetSource(v1alpha1.AWSCodeCommitEventSource(a.awsRegion, a.repo))
	event.SetID(*pullRequest.PullRequestId)
	event.SetData(cloudevents.ApplicationJSON, data)

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
