/*
Copyright (c) 2019 TriggerMesh, Inc

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
	"context"
	"flag"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	log "github.com/sirupsen/logrus"
)

var (
	sink                   string
	repoNameEnv            string
	repoBranchEnv          string
	gitEventsEnv           string
	awsRegionEnv           string
	accountAccessKeyID     string
	accountSecretAccessKey string
	syncTime               = 10
	lastCommit             string
	pullRequestIDs         []*string
)

// PushMessageEvent represent a push message event from codeCommit source
type PushMessageEvent struct {
	Commit           *codecommit.Commit `json:"commit"`
	CommitRepository *string            `json:"commitRepository"`
	CommitBranch     *string            `json:"commitBranch"`
	CommitHash       *string            `json:"commitHash"`
	EventSource      *string            `json:"eventSource"`
	AwsRegion        *string            `json:"awsRegion"`
}

// PRMessageEvent represent a PR message event from codeCommit source
type PRMessageEvent struct {
	PullRequest *codecommit.PullRequest `json:"PullRequest"`
	EventType   *string                 `json:"EventType"`
	Repository  *string                 `json:"Repository"`
	Branch      *string                 `json:"Branch"`
	EventSource *string                 `json:"eventSource"`
	AwsRegion   *string                 `json:"awsRegion"`
}

//Clients struct represent CC Clients
type Clients struct {
	CodeCommit  codecommitiface.CodeCommitAPI
	CloudEvents cloudevents.Client
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")

	//TODO: Make sure all these env vars exist
	repoNameEnv = os.Getenv("REPO")
	repoBranchEnv = os.Getenv("BRANCH")
	gitEventsEnv = os.Getenv("EVENTS")
	awsRegionEnv = os.Getenv("AWS_REGION")
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
}

func main() {

	flag.Parse()

	//Set logging output levels
	_, varPresent := os.LookupEnv("DEBUG")
	if varPresent {
		log.SetLevel(log.DebugLevel)
	}

	//Create client for code commit
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegionEnv),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(sink),
	)
	if err != nil {
		log.Fatal(err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow())
	if err != nil {
		log.Fatal(err)
	}

	clients := Clients{
		CodeCommit:  codecommit.New(sess),
		CloudEvents: c,
	}

	if strings.Contains(gitEventsEnv, "push") {
		log.Info("Push Events Enabled!")

		branchInfo, err := clients.CodeCommit.GetBranch(&codecommit.GetBranchInput{
			BranchName:     aws.String(repoBranchEnv),
			RepositoryName: aws.String(repoNameEnv),
		})
		if err != nil {
			log.Fatal(err)
		}

		lastCommit = *branchInfo.Branch.CommitId

	}

	if strings.Contains(gitEventsEnv, "pull_request") {
		log.Info("Pull Request Events Enabled!")

		//Get pull request IDs
		pullRequestsOutput, err := clients.CodeCommit.ListPullRequests(&codecommit.ListPullRequestsInput{
			RepositoryName: aws.String(repoNameEnv),
		})

		if err != nil {
			log.Fatal(err)
		}

		pullRequestIDs = pullRequestsOutput.PullRequestIds

	}

	if !strings.Contains("gitEventsEnv", "pull_request") && strings.Contains("gitEventsEnv", "push") {
		log.Fatal("error identifying events type. Please, select either `pull_request` or `push` event or both")
	}

	processedPullRequests, err := clients.preparePullRequests()
	if err != nil {
		log.Error(err)
	}

	//range time.Tick(time.Duration(syncTime) * time.Second)
	for {
		if strings.Contains(gitEventsEnv, "push") {
			err := clients.processCommits()
			if err != nil {
				log.Error(err)
			}
		}

		if strings.Contains(gitEventsEnv, "pull_request") {
			pullRequests, err := clients.preparePullRequests()
			if err != nil {
				log.Error(err)
			}

			pullRequests = removeOldPRs(processedPullRequests, pullRequests)

			for _, pr := range pullRequests {

				err = clients.sendPREvent(pr)
				if err != nil {
					log.Error("sendPREvent failed: ", err)
				}
				processedPullRequests = append(processedPullRequests, pr)
			}
		}

	}

}

func (clients Clients) processCommits() error {
	branchInfo, err := clients.CodeCommit.GetBranch(&codecommit.GetBranchInput{
		BranchName:     aws.String(repoBranchEnv),
		RepositoryName: aws.String(repoNameEnv),
	})
	if err != nil {
		return err
	}

	commitOutput, err := clients.CodeCommit.GetCommit(&codecommit.GetCommitInput{
		CommitId:       branchInfo.Branch.CommitId,
		RepositoryName: aws.String(repoNameEnv),
	})
	if err != nil {
		return err
	}

	if *commitOutput.Commit.CommitId == lastCommit {
		return nil
	}

	lastCommit = *commitOutput.Commit.CommitId

	err = clients.sendCommitEvent(commitOutput.Commit)
	if err != nil {
		return err
	}

	return nil
}

func (clients Clients) preparePullRequests() ([]*codecommit.PullRequest, error) {
	pullRequests := []*codecommit.PullRequest{}

	input := codecommit.ListPullRequestsInput{
		RepositoryName: aws.String(repoNameEnv),
	}

	for {
		//Get pull request IDs
		pullRequestsOutput, err := clients.CodeCommit.ListPullRequests(&input)
		if err != nil {
			return pullRequests, err
		}

		prIDs := []*string{}

		prIDs = append(prIDs, pullRequestsOutput.PullRequestIds...)

		for _, id := range prIDs {

			pri := codecommit.GetPullRequestInput{PullRequestId: id}

			prInfo, err := clients.CodeCommit.GetPullRequest(&pri)
			if err != nil {
				return pullRequests, err
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

//sendPush sends an event containing data about a git commit that was pushed to a branch
func (clients Clients) sendCommitEvent(commit *codecommit.Commit) error {

	log.Info("send Commit Event")

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			Type:            "com.amazon.codecommit.commit",
			Subject:         aws.String("AWS CodeCommit Event"),
			Source:          *types.ParseURLRef(""),
			ID:              *commit.CommitId,
			DataContentType: aws.String("application/json"),
		}.AsV03(),
		Data: &PushMessageEvent{
			Commit:           commit,
			CommitRepository: aws.String(repoNameEnv),
			CommitBranch:     aws.String(repoBranchEnv),
			EventSource:      aws.String("aws:codecommit"),
			AwsRegion:        aws.String(awsRegionEnv),
		},
	}

	_, err := clients.CloudEvents.Send(context.Background(), event)
	if err != nil {
		return err
	}

	return nil
}

func (clients Clients) sendPREvent(pullRequest *codecommit.PullRequest) error {
	log.Info("send Pull Request Event")

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			Type:            "com.amazon.codecommit.pull_request",
			Subject:         aws.String("AWS CodeCommit Event"),
			Source:          *types.ParseURLRef(*pullRequest.AuthorArn),
			ID:              *pullRequest.PullRequestId,
			DataContentType: aws.String("application/json"),
		}.AsV03(),
		Data: &PRMessageEvent{
			PullRequest: pullRequest,
			Repository:  aws.String(repoNameEnv),
			Branch:      aws.String(repoBranchEnv),
			EventSource: aws.String("aws:codecommit"),
			AwsRegion:   aws.String(awsRegionEnv),
		},
	}

	_, err := clients.CloudEvents.Send(context.Background(), event)
	if err != nil {
		return err
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
