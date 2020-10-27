/*
Copyright (c) 2019-2020 TriggerMesh Inc.

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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
)

type mockedReceiveMsgs struct {
	sqsiface.SQSAPI
	resp *sqs.ReceiveMessageOutput
	err  error
}

type mockedDeleteMsgs struct {
	sqsiface.SQSAPI
	resp *sqs.DeleteMessageBatchOutput
	err  error
}

type mockedGetQueueURL struct {
	sqsiface.SQSAPI
	resp *sqs.GetQueueUrlOutput
	err  error
}

func (m mockedReceiveMsgs) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return m.resp, m.err
}

func (m mockedDeleteMsgs) DeleteMessageBatch(in *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	return m.resp, m.err
}

func (m mockedGetQueueURL) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) { //nolint:golint,stylecheck
	return m.resp, m.err
}

func TestQueueLookup(t *testing.T) {
	cases := []struct {
		resp     *sqs.GetQueueUrlOutput
		err      error
		expected *string
	}{
		{ // Case 1, expect parsed responses
			resp: &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String("testQueueURL"),
			},
			err:      nil,
			expected: aws.String("testQueueURL"),
		},
		{ // Case 2, expect error
			resp: &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String(""),
			},
			err:      errors.New("fake getQueueUrl error"),
			expected: aws.String(""),
		},
	}

	for _, c := range cases {
		a := &adapter{
			logger: loggingtesting.TestLogger(t),
			sqsClient: mockedGetQueueURL{
				resp: c.resp,
				err:  c.err,
			},
		}

		url, err := a.queueLookup("")
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, url.QueueUrl)
	}
}

func TestGetMessages(t *testing.T) {
	const queueURL = "mockURL"

	cases := []struct {
		resp     *sqs.ReceiveMessageOutput
		err      error
		expected []*sqs.Message
	}{
		{ // Case 1, expect parsed responses
			resp: &sqs.ReceiveMessageOutput{
				Messages: make([]*sqs.Message, 2),
			},
			err:      nil,
			expected: make([]*sqs.Message, 2),
		},
		{ // Case 2, no message returned
			resp:     &sqs.ReceiveMessageOutput{},
			err:      errors.New("no message found"),
			expected: []*sqs.Message{},
		},
	}

	for _, c := range cases {
		a := &adapter{
			logger: loggingtesting.TestLogger(t),
			sqsClient: mockedReceiveMsgs{
				resp: c.resp,
				err:  c.err},
		}

		msgs, err := a.getMessages(queueURL)
		assert.Equal(t, c.err, err)
		assert.Equal(t, len(c.expected), len(msgs))
	}
}

func TestSendEvents(t *testing.T) {
	msgs := []*sqs.Message{
		{
			MessageId: aws.String("0001"),
			Body:      aws.String("msg1"),
		},
		{
			MessageId: aws.String("0002"),
			Body:      aws.String("msg2"),
		},
	}

	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	sent, err := a.sendSQSEvents(msgs)
	ceClientSentEvents := ceClient.Sent()
	assert.NoError(t, err)
	assert.Len(t, sent, len(msgs), "The function didn't return the expected number of messages")
	require.Len(t, ceClientSentEvents, len(msgs), "The client didn't send the expected number of messages")

	sort.Sort(eventsByID(ceClientSentEvents))

	for i, e := range ceClientSentEvents {
		sentMsg := &sqs.Message{}
		err := e.DataAs(sentMsg)
		assert.NoError(t, err)
		assert.EqualValues(t, msgs[i], sentMsg, "%d: sent payload differs from original message", i)
	}
}

type eventsByID []cloudevents.Event

func (ce eventsByID) Len() int           { return len(ce) }
func (ce eventsByID) Less(i, j int) bool { return ce[i].ID() < ce[j].ID() }
func (ce eventsByID) Swap(i, j int)      { ce[i], ce[j] = ce[j], ce[i] }

func TestDeleteMessage(t *testing.T) {
	const queueURL = "mockURL"

	msgs := []*sqs.Message{
		{
			MessageId:     aws.String("0001"),
			ReceiptHandle: aws.String("0001"),
		},
		{
			MessageId:     aws.String("0002"),
			ReceiptHandle: aws.String("0002"),
		},
	}

	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	a.sqsClient = mockedDeleteMsgs{
		resp: &sqs.DeleteMessageBatchOutput{},
		err:  nil,
	}

	err := a.deleteMessages(queueURL, msgs)
	assert.NoError(t, err)

	a.sqsClient = mockedDeleteMsgs{
		resp: &sqs.DeleteMessageBatchOutput{},
		err:  errors.New("fake deleteMessage error"),
	}

	err = a.deleteMessages(queueURL, msgs)
	assert.Error(t, err)
}

func TestErrList(t *testing.T) {
	testCases := []struct {
		name   string
		errors []error
		expect string
	}{
		{
			name:   "no error",
			errors: nil,
			expect: "",
		},
		{
			name: "one error",
			errors: []error{
				errors.New("err1"),
			},
			expect: "err1",
		},
		{
			name: "multiple errors",
			errors: []error{
				errors.New("err1"),
				errors.New("err2"),
				errors.New("err3"),
			},
			expect: "[err1, err2, err3]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errs := &errList{errs: tc.errors}
			assert.EqualError(t, errs, tc.expect)
		})
	}
}
