package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
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

type codecommitMsg struct{}

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
	ccClient := codecommit.New(sess)
	log.Info("Started new session")
	m := codecommitMsg{}
	m.ReceiveMsg(ccClient)

}

//ReceiveMsg implements the receive interface for codecommit
func (codecommitMsg) ReceiveMsg(ccClient *codecommit.CodeCommit) {
	log.Info("Started receiving messages")
	//Parse out git events we're looking for
	gitEvents := strings.Split(gitEventsEnv, ",")

	var gitCommit string
	var pullRequests map[string]*codecommit.PullRequest
	var err error

	//Seed commit and PRs
	for _, gitEvent := range gitEvents {
		switch gitEvent {

		case "push":
			gitCommit, err = commitID(ccClient)
			if err != nil {
				log.Fatal("Unable to seed commits: ", err)
			}

		case "pull_request":
			pullRequests = make(map[string]*codecommit.PullRequest)
			pullRequestsOutput, err := ccClient.ListPullRequests(&codecommit.ListPullRequestsInput{
				RepositoryName: aws.String(repoNameEnv)})
			if err != nil {
				log.Fatal("Unable to seed PRs: ", err)
			}
			for _, pr := range aws.StringValueSlice(pullRequestsOutput.PullRequestIds) {
				err = appendPR(ccClient, pr, &pullRequests)
				if err != nil {
					log.Fatal("Unable to seed PRs: ", err)
				}
			}
		}
	}

	//Look for new messages every x seconds
	for range time.Tick(time.Duration(syncTime) * time.Second) {

		for _, gitEvent := range gitEvents {
			switch gitEvent {

			//If push in events, get last commit ID. Send event if it's changed since last time.
			case "push":
				gitCommitTemp, err := commitID(ccClient)
				if err != nil {
					log.Fatal(err)
				}
				if gitCommitTemp != gitCommit {
					gitCommit = gitCommitTemp
					err = sendPushEvent(ccClient, gitCommit, sink)
					if err != nil {
						log.Error("Failed to send push event. ", err)
					}
				}

			//If PR in events, fetch PRs and push msg if necessary
			case "pull_request":
				//Get pull request IDs
				pullRequestsOutput, err := ccClient.ListPullRequests(&codecommit.ListPullRequestsInput{
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
						prInfo, err := ccClient.GetPullRequest(&codecommit.GetPullRequestInput{
							PullRequestId: aws.String(pr),
						})
						if err != nil {
							log.Error(err)
						}
						if localStatus != aws.StringValue(prInfo.PullRequest.PullRequestStatus) {
							pullRequests[pr] = prInfo.PullRequest
							err = sendPREvent(ccClient, pullRequests[pr], "pr_"+strings.ToLower(aws.StringValue(pullRequests[pr].PullRequestStatus)), sink)
							if err != nil {
								log.Error("Error sending PR event: ", err)
							}
						}
						// If we don't know about this PR, assume it's new and pr_open event
					} else {
						appendPR(ccClient, pr, &pullRequests)
						err = sendPREvent(ccClient, pullRequests[pr], "pr_open", sink)
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
func appendPR(ccClient *codecommit.CodeCommit, prID string, prList *map[string]*codecommit.PullRequest) error {
	prInfo, err := ccClient.GetPullRequest(&codecommit.GetPullRequestInput{
		PullRequestId: aws.String(prID),
	})
	if err != nil {
		return err
	}

	(*prList)[prID] = prInfo.PullRequest
	return nil
}

//sendPREvent sends an event contianing PR info when a PR is open/closed
func sendPREvent(ccClient *codecommit.CodeCommit, pr *codecommit.PullRequest, eventType, sink string) error {
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

//commitID returns latest commit hash on the branch
func commitID(ccClient *codecommit.CodeCommit) (string, error) {
	branchInfo, err := ccClient.GetBranch(&codecommit.GetBranchInput{
		BranchName:     aws.String(repoBranchEnv),
		RepositoryName: aws.String(repoNameEnv)})
	if err != nil {
		return "", err
	}

	return aws.StringValue(branchInfo.Branch.CommitId), nil
}

//sendPush sends an event containing data about a git commit that was pushed to a branch
func sendPushEvent(ccClient *codecommit.CodeCommit, commitHash, sink string) error {

	//Fetch full commit info
	commitOutput, err := ccClient.GetCommit(&codecommit.GetCommitInput{
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
