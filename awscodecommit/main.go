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
	"flag"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	"github.com/knative/pkg/cloudevents"
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

//СodeCommitClient struct represent CC Client
type СodeCommitClient struct {
	Client            codecommitiface.CodeCommitAPI
	CloudEventsClient cloudevents.Client
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

	cloudEvents := cloudevents.NewClient(
		sink,
		cloudevents.Builder{
			Source:    "aws:codecommit",
			EventType: "codecommit event",
		},
	)

	cc := СodeCommitClient{
		Client:            codecommit.New(sess),
		CloudEventsClient: *cloudEvents,
	}

	if strings.Contains(gitEventsEnv, "push") {
		log.Info("Push Events Enabled!")

		branchInfo, err := cc.Client.GetBranch(&codecommit.GetBranchInput{
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
		pullRequestsOutput, err := cc.Client.ListPullRequests(&codecommit.ListPullRequestsInput{
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

	//range time.Tick(time.Duration(syncTime) * time.Second)
	for {
		if strings.Contains(gitEventsEnv, "push") {
			err := cc.processCommits()
			if err != nil {
				log.Error(err)
			}
		}

		if strings.Contains(gitEventsEnv, "pull_request") {
			err = cc.processPullRequest()
			if err != nil {
				log.Error(err)
			}
		}

	}

}

func (cc СodeCommitClient) processCommits() error {
	branchInfo, err := cc.Client.GetBranch(&codecommit.GetBranchInput{
		BranchName:     aws.String(repoBranchEnv),
		RepositoryName: aws.String(repoNameEnv),
	})
	if err != nil {
		return err
	}

	commitOutput, err := cc.Client.GetCommit(&codecommit.GetCommitInput{
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

	err = cc.sendCommitEvent(commitOutput.Commit)
	if err != nil {
		return err
	}

	return nil
}

func (cc СodeCommitClient) processPullRequest() error {
	prIDs := []*string{}

	input := codecommit.ListPullRequestsInput{
		RepositoryName: aws.String(repoNameEnv),
	}

	for {
		//Get pull request IDs
		pullRequestsOutput, err := cc.Client.ListPullRequests(&input)

		if err != nil {
			return err
		}

		prIDs = append(prIDs, pullRequestsOutput.PullRequestIds...)

		if pullRequestsOutput.NextToken == nil {
			break
		}

		input.NextToken = pullRequestsOutput.NextToken
	}

	for _, id := range prIDs {
		log.Infof("Process ID [%v]", *id)
		if contains(pullRequestIDs, *id) {
			continue
		}
		pullRequestIDs = append(pullRequestIDs, id)

		pri := codecommit.GetPullRequestInput{PullRequestId: id}

		log.Info(pri)

		prInfo, err := cc.Client.GetPullRequest(&pri)
		if err != nil {
			return err
		}
		err = cc.sendPREvent(prInfo.PullRequest)
		if err != nil {
			return err
		}
	}

	return nil
}

//sendPush sends an event containing data about a git commit that was pushed to a branch
func (cc СodeCommitClient) sendCommitEvent(commit *codecommit.Commit) error {
	log.Info("send Commit Event")

	codecommitEvent := PushMessageEvent{
		Commit:           commit,
		CommitRepository: aws.String(repoNameEnv),
		CommitBranch:     aws.String(repoBranchEnv),
		EventSource:      aws.String("aws:codecommit"),
		AwsRegion:        aws.String(awsRegionEnv),
	}

	if err := cc.CloudEventsClient.Send(codecommitEvent); err != nil {
		return err
	}

	return nil
}

func (cc СodeCommitClient) sendPREvent(pullRequest *codecommit.PullRequest) error {
	log.Info("send Pull Request Event")

	codecommitEvent := PRMessageEvent{
		PullRequest: pullRequest,
		Repository:  aws.String(repoNameEnv),
		Branch:      aws.String(repoBranchEnv),
		EventSource: aws.String("aws:codecommit"),
		AwsRegion:   aws.String(awsRegionEnv),
	}

	if err := cc.CloudEventsClient.Send(codecommitEvent); err != nil {
		return err
	}

	return nil
}

// Contains tells whether a contains x.
func contains(a []*string, x string) bool {
	for _, n := range a {
		if x == *n {
			return true
		}
	}
	return false
}
