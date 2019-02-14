package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	"github.com/jarcoal/httpmock"
	"github.com/knative/pkg/cloudevents"
	"github.com/stretchr/testify/assert"
)

type mockedClientForCommits struct {
	codecommitiface.CodeCommitAPI
	GetBranchResp codecommit.GetBranchOutput
	GetCommitResp codecommit.GetCommitOutput
	GetBranchErr  error
	GetCommitErr  error
}

type mockedClientForPRs struct {
	codecommitiface.CodeCommitAPI
	ListPRsResp codecommit.ListPullRequestsOutput
	GetPRResp   codecommit.GetPullRequestOutput
	ListPRsErr  error
	GetPRErr    error
}

func (m mockedClientForCommits) GetBranch(in *codecommit.GetBranchInput) (*codecommit.GetBranchOutput, error) {
	return &m.GetBranchResp, m.GetBranchErr
}

func (m mockedClientForCommits) GetCommit(in *codecommit.GetCommitInput) (*codecommit.GetCommitOutput, error) {
	return &m.GetCommitResp, m.GetCommitErr
}

func (m mockedClientForPRs) ListPullRequests(in *codecommit.ListPullRequestsInput) (*codecommit.ListPullRequestsOutput, error) {
	return &m.ListPRsResp, m.ListPRsErr
}

func (m mockedClientForPRs) GetPullRequests(in *codecommit.GetPullRequestInput) (*codecommit.GetPullRequestOutput, error) {
	return &m.GetPRResp, m.GetPRErr
}

func TestSendPREvent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	client := СodeCommitClient{
		CloudEventsClient: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:codecommit",
				EventType: "codecommit event",
			},
		),
	}

	pullRequest := codecommit.PullRequest{}

	err := client.sendPREvent(&pullRequest)
	assert.Error(t, err)

	c := cloudevents.NewClient(
		"https://foo.com",
		cloudevents.Builder{
			Source:    "aws:codecommit",
			EventType: "codecommit event",
		},
	)

	client.CloudEventsClient = c

	err = client.sendPREvent(&pullRequest)
	assert.NoError(t, err)
}

func TestSendCommitEvent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	client := СodeCommitClient{
		CloudEventsClient: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:codecommit",
				EventType: "codecommit event",
			},
		),
	}

	commit := codecommit.Commit{}
	err := client.sendCommitEvent(&commit)
	assert.Error(t, err)

	c := cloudevents.NewClient(
		"https://foo.com",
		cloudevents.Builder{
			Source:    "aws:codecommit",
			EventType: "codecommit event",
		},
	)

	client.CloudEventsClient = c

	err = client.sendCommitEvent(&commit)
	assert.NoError(t, err)
}

func TestProcessCommits(t *testing.T) {

}

func TestProcessPullRequest(t *testing.T) {

}
