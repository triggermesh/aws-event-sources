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

package awssqssource

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/assert"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
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

type mockedGetQueueUrl struct {
	sqsiface.SQSAPI
	Resp sqs.GetQueueUrlOutput
	err  error
}

func (m mockedReceiveMsgs) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return &m.Resp, m.err
}

func (m mockedDeleteMsgs) DeleteMessage(in *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return &m.Resp, m.err
}

func (m mockedGetQueueUrl) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return &m.Resp, m.err
}

func TestQueueLookup(t *testing.T) {
	cases := []struct {
		Resp     sqs.GetQueueUrlOutput
		err      error
		Expected *string
	}{
		{ // Case 1, expect parsed responses
			Resp:     sqs.GetQueueUrlOutput{QueueUrl: aws.String("testQueueURL")},
			err:      nil,
			Expected: aws.String("testQueueURL"),
		},
		{ // Case 2, expect error
			Resp:     sqs.GetQueueUrlOutput{QueueUrl: aws.String("")},
			err:      errors.New("fake getQueueUrl error"),
			Expected: aws.String(""),
		},
	}

	for _, c := range cases {
		a := &adapter{
			logger:    loggingtesting.TestLogger(t),
			sqsClient: mockedGetQueueUrl{Resp: c.Resp, err: c.err},
		}

		url, err := a.queueLookup("testQueue")
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.Expected, url.QueueUrl)
	}
}

func TestGetMessages(t *testing.T) {
	const queueURL = "mockURL"

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
		a := &adapter{
			logger:    loggingtesting.TestLogger(t),
			sqsClient: mockedReceiveMsgs{Resp: c.Resp, err: c.err},
		}

		msgs, err := a.getMessages(queueURL, waitTimeoutSec)
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

	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	err := a.sendSQSEvent(&msg, aws.String("testQueueARN"))
	assert.NoError(t, err)

	gotEvents := ceClient.Sent()
	assert.Len(t, gotEvents, 1, "Expected 1 event, got %d", len(gotEvents))

	wantData := `{"messageId":"foo","receiptHandle":null,"body":"bar","attributes":{"SentTimestamp":"1549540781"},"messageAttributes":null,"md5OfBody":null,"eventSource":"aws:sqs","eventSourceARN":"testQueueARN","awsRegion":""}`
	gotData := string(gotEvents[0].Data())
	assert.EqualValues(t, wantData, gotData, "Expected event %q, got %q", wantData, gotData)
}

func TestDeleteMessage(t *testing.T) {
	const queueURL = "mockURL"

	msg := sqs.Message{
		ReceiptHandle: aws.String("foo"),
	}

	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	a.sqsClient = mockedDeleteMsgs{
		Resp: sqs.DeleteMessageOutput{},
		err:  nil,
	}

	err := a.deleteMessage(queueURL, msg.ReceiptHandle)
	assert.NoError(t, err)

	a.sqsClient = mockedDeleteMsgs{
		Resp: sqs.DeleteMessageOutput{},
		err:  errors.New("fake deleteMessage error"),
	}

	err = a.deleteMessage(queueURL, msg.ReceiptHandle)
	assert.Error(t, err)
}
