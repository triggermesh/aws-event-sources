package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type mockedReceiveMsgs struct {
	sqsiface.SQSAPI
	Resp sqs.ReceiveMessageOutput
}

type mockedDeleteMsgs struct {
	sqsiface.SQSAPI
	Resp sqs.DeleteMessageOutput
}

type mockedListQueues struct {
	sqsiface.SQSAPI
	Resp sqs.ListQueuesOutput
}

func (m mockedReceiveMsgs) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return &m.Resp, nil
}

func (m mockedDeleteMsgs) DeleteMessage(in *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return &m.Resp, nil
}

func (m mockedListQueues) ListQueues(in *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	return &m.Resp, nil
}

func TestQueueLookup(t *testing.T) {
	cases := []struct {
		Resp     sqs.ListQueuesOutput
		Expected string
	}{
		{ // Case 1, expect parsed responses
			Resp: sqs.ListQueuesOutput{
				QueueUrls: []*string{aws.String("testQueueURL")}},
			Expected: "testQueueURL",
		},
		{ // Case 1, expect parsed responses
			Resp: sqs.ListQueuesOutput{
				QueueUrls: []*string{aws.String("testQueueURL"), aws.String("testQueueURLFoo")}},
			Expected: "testQueueURL",
		},
	}

	for _, c := range cases {
		q := Queue{
			Client: mockedListQueues{Resp: c.Resp},
		}
		url, err := q.QueueLookup("testQueue")
		assert.NoError(t, err)
		assert.Equal(t, c.Expected, url)
	}
}

func TestGetMessages(t *testing.T) {
	cases := []struct {
		Resp     sqs.ReceiveMessageOutput
		Expected []sqs.Message
	}{
		{ // Case 1, expect parsed responses
			Resp: sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{},
				},
			},
			Expected: []sqs.Message{
				{},
			},
		},
		{ // Case 2, not messages returned
			Resp:     sqs.ReceiveMessageOutput{},
			Expected: []sqs.Message{},
		},
	}

	for _, c := range cases {
		q := Queue{
			Client: mockedReceiveMsgs{Resp: c.Resp},
			URL:    aws.String("mockURL"),
		}
		msgs, err := q.GetMessages(20)
		assert.NoError(t, err)
		assert.Equal(t, len(c.Expected), len(msgs))
	}
}

func TestPushMessage(t *testing.T) {
	msg := sqs.Message{}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	err := pushMessage(&msg, "")
	assert.Error(t, err)

	msg = sqs.Message{
		MessageId: aws.String("foo"),
		Body:      aws.String("bar"),
		Attributes: map[string]*string{
			"SentTimestamp": aws.String("1549540781"),
		},
	}
	dryRun = true
	err = pushMessage(&msg, "https://foo.com")
	assert.NoError(t, err)
}

func TestDeleteMessage(t *testing.T) {

	q := Queue{
		Client: mockedDeleteMsgs{Resp: sqs.DeleteMessageOutput{}},
		URL:    aws.String("mockURL"),
	}

	msg := sqs.Message{
		ReceiptHandle: aws.String("foo"),
	}
	err := q.DeleteMessage(msg.ReceiptHandle)
	assert.NoError(t, err)
}
