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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/cloudevents/sdk-go"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type mockedReceiveMsgs struct {
	sqsiface.SQSAPI
	Resp sqs.ReceiveMessageOutput
	err  error
}

type mockedDeleteMsgs struct {
	sqsiface.SQSAPI
	Resp sqs.DeleteMessageOutput
	err  error
}

type mockedListQueues struct {
	sqsiface.SQSAPI
	Resp sqs.ListQueuesOutput
	err  error
}

func (m mockedReceiveMsgs) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return &m.Resp, m.err
}

func (m mockedDeleteMsgs) DeleteMessage(in *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return &m.Resp, m.err
}

func (m mockedListQueues) ListQueues(in *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	return &m.Resp, m.err
}

func TestQueueLookup(t *testing.T) {
	cases := []struct {
		Resp     sqs.ListQueuesOutput
		err      error
		Expected string
	}{
		{ // Case 1, expect parsed responses
			Resp: sqs.ListQueuesOutput{
				QueueUrls: []*string{aws.String("testQueueURL")}},
			err:      nil,
			Expected: "testQueueURL",
		},
		{ // Case 2, expect parsed responses
			Resp: sqs.ListQueuesOutput{
				QueueUrls: []*string{aws.String("testQueueURL"), aws.String("testQueueURLFoo")}},
			err:      nil,
			Expected: "testQueueURL",
		},
		{ // Case 3, expect error
			Resp: sqs.ListQueuesOutput{
				QueueUrls: []*string{}},
			err:      errors.New("No such queue"),
			Expected: "",
		},
	}

	for _, c := range cases {
		clients := Clients{
			SQS: mockedListQueues{Resp: c.Resp, err: c.err},
		}
		url, err := clients.QueueLookup("testQueue")
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.Expected, url)
	}
}

func TestGetMessages(t *testing.T) {
	cases := []struct {
		Resp     sqs.ReceiveMessageOutput
		err      error
		Expected []sqs.Message
	}{
		{ // Case 1, expect parsed responses
			Resp: sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{},
				},
			},
			err: nil,
			Expected: []sqs.Message{
				{},
			},
		},
		{ // Case 2, not messages returned
			Resp:     sqs.ReceiveMessageOutput{},
			err:      errors.New("No messages found"),
			Expected: []sqs.Message{},
		},
	}

	for _, c := range cases {
		clients := Clients{
			SQS: mockedReceiveMsgs{Resp: c.Resp, err: c.err},
		}
		queueURL = "mockURL"
		msgs, err := clients.GetMessages(20)
		assert.Equal(t, c.err, err)
		assert.Equal(t, len(c.Expected), len(msgs))
	}
}

func TestPushMessage(t *testing.T) {
	msg := sqs.Message{
		MessageId: aws.String("foo"),
		Body:      aws.String("bar"),
		Attributes: map[string]*string{
			"SentTimestamp": aws.String("1549540781"),
		}}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget("https://foo.com"),
	)
	assert.NoError(t, err)

	cloudClient, err := cloudevents.NewClient(transport)
	assert.NoError(t, err)

	clients := Clients{
		CloudEvents: cloudClient,
	}

	err = clients.sendSQSEvent(&msg, aws.String("testQueueARN"))
	assert.NoError(t, err)
}

func TestDeleteMessage(t *testing.T) {

	clients := Clients{
		SQS: mockedDeleteMsgs{Resp: sqs.DeleteMessageOutput{}, err: nil},
	}

	msg := sqs.Message{
		ReceiptHandle: aws.String("foo"),
	}
	err := clients.DeleteMessage(msg.ReceiptHandle)
	assert.NoError(t, err)

	clients = Clients{
		SQS: mockedDeleteMsgs{Resp: sqs.DeleteMessageOutput{}, err: errors.New("Could not delete msg")},
	}

	msg = sqs.Message{
		ReceiptHandle: aws.String("foo"),
	}
	err = clients.DeleteMessage(msg.ReceiptHandle)
	assert.Error(t, err)
}
