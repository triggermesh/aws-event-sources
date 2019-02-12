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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/knative/pkg/cloudevents"
	log "github.com/sirupsen/logrus"
	"github.com/triggermesh/sources/tmevents"
)

var (
	sink                   string
	repoNameEnv            string
	repoBranchEnv          string
	gitEventsEnv           string
	channelEnv             string
	namespaceEnv           string
	awsRegionEnv           string
	accountAccessKeyID     string
	accountSecretAccessKey string
	syncTime               = 10
)

//СodeCommitClient struct represent CC Client
type СodeCommitClient struct {
	Client            *codecommit.CodeCommit
	CloudEventsClient *cloudevents.Client
	GitEvents         []string
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	//Set logging output levels
	_, varPresent := os.LookupEnv("DEBUG")
	if varPresent {
		log.SetLevel(log.DebugLevel)
	}

	//TODO: Make sure all these env vars exist
	repoNameEnv = os.Getenv("REPO")
	repoBranchEnv = os.Getenv("BRANCH")
	gitEventsEnv = os.Getenv("EVENTS")
	awsRegionEnv = os.Getenv("AWS_REGION")
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

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
			Source: "aws:codecommit",
		},
	)

	cc := СodeCommitClient{
		Client:            codecommit.New(sess),
		CloudEventsClient: cloudEvents,
		GitEvents:         strings.Split(gitEventsEnv, ","),
	}

	gitCommit, pullRequests, err := cc.SeedCommitsAndPRs()
	if err != nil {
		log.Fatal(err)
	}

	cc.ReceiveMsg(gitCommit, pullRequests)
}

//SeedCommitsAndPRs prepares commit and PRs
func (cc *СodeCommitClient) SeedCommitsAndPRs() (gitCommit string, pullRequests map[string]*codecommit.PullRequest, err error) {
	log.Info("Started receiving messages")

	for _, gitEvent := range cc.GitEvents {
		switch gitEvent {

		case "push":
			gitCommit, err = cc.getCommitID()
			if err != nil {
				return gitCommit, pullRequests, err
			}

		case "pull_request":
			pullRequests = make(map[string]*codecommit.PullRequest)
			pullRequestsOutput, err := cc.Client.ListPullRequests(&codecommit.ListPullRequestsInput{
				RepositoryName: aws.String(repoNameEnv),
			})
			if err != nil {
				return gitCommit, pullRequests, err
			}

			for _, pr := range aws.StringValueSlice(pullRequestsOutput.PullRequestIds) {
				err = cc.appendPR(pr, &pullRequests)
				if err != nil {
					return gitCommit, pullRequests, err
				}
			}
		default:
			return gitCommit, pullRequests, fmt.Errorf("unexpected git event %s", gitEvent)
		}
	}

	return gitCommit, pullRequests, nil
}

//ReceiveMsg implements the receive interface for codecommit
func (cc *СodeCommitClient) ReceiveMsg(gitCommit string, pullRequests map[string]*codecommit.PullRequest) {

	//Look for new messages every x seconds
	for range time.Tick(time.Duration(syncTime) * time.Second) {

		for _, gitEvent := range cc.GitEvents {
			switch gitEvent {

			//If push in events, get last commit ID. Send event if it's changed since last time.
			case "push":
				gitCommitTemp, err := cc.getCommitID()
				if err != nil {
					log.Fatal(err)
				}
				if gitCommitTemp != gitCommit {
					gitCommit = gitCommitTemp
					err = cc.sendPushEvent(gitCommit, sink)
					if err != nil {
						log.Error("Failed to send push event. ", err)
					}
				}

			//If PR in events, fetch PRs and push msg if necessary
			case "pull_request":
				//Get pull request IDs
				pullRequestsOutput, err := cc.Client.ListPullRequests(&codecommit.ListPullRequestsInput{
					RepositoryName: aws.String(repoNameEnv)})
				if err != nil {
					log.Error("Unable to pull PRs: ", err)
					break
				}
				//Check if we already know about the PR ID
				for _, pr := range aws.StringValueSlice(pullRequestsOutput.PullRequestIds) {
					_, ok := pullRequests[pr]
					//If we already know about it, check if statuses match. Send event if not.
					if ok {
						localStatus := aws.StringValue(pullRequests[pr].PullRequestStatus)
						prInfo, err := cc.Client.GetPullRequest(&codecommit.GetPullRequestInput{
							PullRequestId: aws.String(pr),
						})
						if err != nil {
							log.Error(err)
						}
						if localStatus != aws.StringValue(prInfo.PullRequest.PullRequestStatus) {
							pullRequests[pr] = prInfo.PullRequest
							err = cc.sendPREvent(pullRequests[pr], "pr_"+strings.ToLower(aws.StringValue(pullRequests[pr].PullRequestStatus)), sink)
							if err != nil {
								log.Error("Error sending PR event: ", err)
							}
						}
						// If we don't know about this PR, assume it's new and pr_open event
					} else {
						cc.appendPR(pr, &pullRequests)
						err = cc.sendPREvent(pullRequests[pr], "pr_open", sink)
						if err != nil {
							log.Error("Error sending PR event: ", err)
						}
					}
				}
			}
		}
	}
}

//appendPR is here to add PRs to a list we've gotta keep up with so we can see when they are added and closed
func (cc *СodeCommitClient) appendPR(prID string, prList *map[string]*codecommit.PullRequest) error {
	prInfo, err := cc.Client.GetPullRequest(&codecommit.GetPullRequestInput{
		PullRequestId: aws.String(prID),
	})
	if err != nil {
		return err
	}

	(*prList)[prID] = prInfo.PullRequest
	return nil
}

//commitID returns latest commit hash on the branch
func (cc СodeCommitClient) getCommitID() (string, error) {
	branchInfo, err := cc.Client.GetBranch(&codecommit.GetBranchInput{
		BranchName:     aws.String(repoBranchEnv),
		RepositoryName: aws.String(repoNameEnv)})
	if err != nil {
		return "", err
	}

	return aws.StringValue(branchInfo.Branch.CommitId), nil
}

//sendPREvent sends an event contianing PR info when a PR is open/closed
func (cc *СodeCommitClient) sendPREvent(pr *codecommit.PullRequest, eventType, sink string) error {

	eventInfo := tmevents.EventInfo{
		EventData:   []byte(aws.StringValue(pr.Title)),
		EventID:     aws.StringValue(pr.PullRequestId),
		EventTime:   time.Now(),
		EventType:   eventType,
		EventSource: "codecommit",
	}

	log.Debug(eventInfo)

	err := tmevents.PushEvent(&eventInfo, sink)
	if err != nil {
		return err
	}

	return nil
}

//sendPush sends an event containing data about a git commit that was pushed to a branch
func (cc СodeCommitClient) sendPushEvent(commitHash, sink string) error {

	//Fetch full commit info
	commitOutput, err := cc.Client.GetCommit(&codecommit.GetCommitInput{
		CommitId:       aws.String(commitHash),
		RepositoryName: aws.String(repoNameEnv)})
	if err != nil {
		return err
	}
	commit := commitOutput.Commit

	eventInfo := tmevents.EventInfo{
		EventData:   []byte(aws.StringValue(commit.Message)),
		EventID:     aws.StringValue(commit.CommitId),
		EventTime:   time.Now(),
		EventType:   "push",
		EventSource: "codecommit",
	}

	log.Debug(eventInfo)

	err = tmevents.PushEvent(&eventInfo, sink)
	if err != nil {
		return err
	}

	return nil
}
