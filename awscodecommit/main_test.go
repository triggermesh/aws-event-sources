package main

import (
	"errors"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
		CloudEventsClient: *cloudevents.NewClient(
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

	client.CloudEventsClient = *c

	err = client.sendPREvent(&pullRequest)
	assert.NoError(t, err)
}

func TestSendCommitEvent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	client := СodeCommitClient{
		CloudEventsClient: *cloudevents.NewClient(
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

	client.CloudEventsClient = *c

	err = client.sendCommitEvent(&commit)
	assert.NoError(t, err)
}

func TestProcessCommits(t *testing.T) {
	lastCommit = "foo"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	testCases := []struct {
		GetBranchResp codecommit.GetBranchOutput
		GetCommitResp codecommit.GetCommitOutput
		GetBranchErr  error
		GetCommitErr  error
		Sink          string
		Err           error
	}{
		{
			GetBranchResp: codecommit.GetBranchOutput{},
			GetBranchErr:  errors.New("get branch failed"),
			Err:           errors.New("get branch failed"),
		},
		{
			GetBranchResp: codecommit.GetBranchOutput{
				Branch: &codecommit.BranchInfo{CommitId: aws.String("123")},
			},
			GetCommitResp: codecommit.GetCommitOutput{},
			GetBranchErr:  nil,
			GetCommitErr:  errors.New("get commit failed"),
			Err:           errors.New("get commit failed"),
		},
		{
			GetBranchResp: codecommit.GetBranchOutput{
				Branch: &codecommit.BranchInfo{CommitId: aws.String("123")},
			},
			GetCommitResp: codecommit.GetCommitOutput{Commit: &codecommit.Commit{CommitId: aws.String("foo")}},
			GetBranchErr:  nil,
			GetCommitErr:  nil,
			Err:           nil,
		},
		{
			GetBranchResp: codecommit.GetBranchOutput{
				Branch: &codecommit.BranchInfo{CommitId: aws.String("123")},
			},
			GetCommitResp: codecommit.GetCommitOutput{Commit: &codecommit.Commit{CommitId: aws.String("bar")}},
			GetBranchErr:  nil,
			GetCommitErr:  nil,
			Err:           &url.Error{"Post", "", errors.New("no responder found")},
		},
		{
			GetBranchResp: codecommit.GetBranchOutput{
				Branch: &codecommit.BranchInfo{CommitId: aws.String("123")},
			},
			GetCommitResp: codecommit.GetCommitOutput{Commit: &codecommit.Commit{CommitId: aws.String("bar")}},
			GetBranchErr:  nil,
			GetCommitErr:  nil,
			Sink:          "https://foo.com",
			Err:           nil,
		},
	}

	for _, tt := range testCases {

		client := СodeCommitClient{
			Client: mockedClientForCommits{
				GetBranchResp: tt.GetBranchResp,
				GetCommitResp: tt.GetCommitResp,
				GetBranchErr:  tt.GetBranchErr,
				GetCommitErr:  tt.GetCommitErr,
			},
			CloudEventsClient: *cloudevents.NewClient(
				tt.Sink,
				cloudevents.Builder{
					Source:    "aws:codecommit",
					EventType: "codecommit event",
				},
			),
		}

		err := client.processCommits()
		assert.Equal(t, tt.Err, err)

	}

}

func TestProcessPullRequest(t *testing.T) {

}

func TestContains(t *testing.T) {
	testCases := []struct {
		slice   []*string
		element string
		result  bool
	}{
		{[]*string{aws.String("1"), aws.String("2"), aws.String("3")}, "1", true},
		{[]*string{aws.String("1"), aws.String("2"), aws.String("3")}, "4", false},
	}

	for _, tt := range testCases {
		r := contains(tt.slice, tt.element)
		assert.Equal(t, tt.result, r)
	}
}
