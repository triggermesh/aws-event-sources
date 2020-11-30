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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
)

const (
	tQueueArnResource = "MyQueue"
	tQueueURL         = "https://sqs.us-fake-0.amazonaws.com/123456789012/MyQueue"
	tMsgIDPrefix      = "00000000-0000-0000-0000-000000000" // + 3 digits appended for each msg
	tSenderID         = "test"
)

func TestAdapter(t *testing.T) {
	// The test's data is pre-populated so the flow of records is
	// uninterrupted until every message has been retrieved. We can
	// therefore affirm something went wrong if the receiveMsgPeriod timer
	// happens during a test.
	const testTimeout = receiveMsgPeriod

	testCases := map[string]struct {
		numMsgs         int
		queueBufSize    int
		deleteMsgPeriod time.Duration
		expectDeletes   int
	}{
		/* This test ensures a deletion of processed messages
		   is triggered at the end of deleteMsgPeriod.

		   numMsgs is deliberately set to a value that is not a
		   multiple of queueBufSize.
		*/
		"scheduled deletion": {
			numMsgs:         maxDeleteMsgBatchSize + 1,
			queueBufSize:    maxDeleteMsgBatchSize,
			deleteMsgPeriod: 10 * time.Millisecond,
			expectDeletes:   2, // 1 due to bufSize, 1 on timer expiration
		},

		/* These tests ensure our implementation isn't reliant on
		   specific buffer sizes.

		   deleteMsgPeriod is purposely set to a high value. All test
		   values assume messages are received in multiples of
		   maxDeleteMsgBatchSize.
		*/
		"no queue buffer": {
			numMsgs:         maxDeleteMsgBatchSize,
			queueBufSize:    0,
			deleteMsgPeriod: testTimeout * 2,
			expectDeletes:   1,
		},
		"small queue buffers": {
			numMsgs:         maxDeleteMsgBatchSize * 2,
			queueBufSize:    1,
			deleteMsgPeriod: testTimeout * 2,
			expectDeletes:   2,
		},
		"large queue buffers": {
			numMsgs:         maxDeleteMsgBatchSize * 3,
			queueBufSize:    100,
			deleteMsgPeriod: testTimeout * 2,
			expectDeletes:   3,
		},
	}

	for name, tc := range testCases {
		//nolint:scopelint
		t.Run(name, func(t *testing.T) {
			ceClient := adaptertest.NewTestClient()

			sqsClient := &standardMockSQSClient{
				availMsgs: makeMockMessages(tc.numMsgs),
			}

			a := adapter{
				logger: loggingtesting.TestLogger(t),

				sqsClient: sqsClient,
				ceClient:  ceClient,

				arn: makeARN(tQueueArnResource),

				processQueue: make(chan *sqs.Message, tc.queueBufSize),
				deleteQueue:  make(chan *sqs.Message, tc.queueBufSize),

				deletePeriod: tc.deleteMsgPeriod,
			}

			testCtx, testCancel := context.WithTimeout(context.Background(), testTimeout)
			defer testCancel()

			startCtx, startCancel := context.WithCancel(testCtx)
			defer startCancel()

			errCh := make(chan error)
			defer close(errCh)

			go func() {
				errCh <- a.Start(startCtx)
			}()

			timer := time.NewTimer(0)
			defer timer.Stop()

		pollSent:
			for {
				select {
				case <-testCtx.Done():
					assert.NoError(t, <-errCh)
					t.Fatal("Timeout waiting for events")

				case <-timer.C:
					if len(ceClient.Sent()) >= tc.numMsgs {
						break pollSent
					}
					timer.Reset(5 * time.Millisecond)
				}
			}

			timer.Reset(0)

		pollInFlight:
			for {
				select {
				case <-testCtx.Done():
					assert.NoError(t, <-errCh)
					t.Fatal("Timeout waiting for in-flight messages to be deleted")

				case <-timer.C:
					if sqsClient.callsDelete >= tc.expectDeletes {
						startCancel()
						break pollInFlight
					}
					timer.Reset(5 * time.Millisecond)
				}
			}

			// no matter what, Start() should always return after its context has
			// been cancelled
			select {
			case <-testCtx.Done():
				assert.NoError(t, <-errCh)
				t.Fatal("Timeout waiting for Start to return")

			case err := <-errCh:
				assert.NoError(t, err)
			}

			// asserting a single event suffices since the entire data set is mocked
			ev := ceClient.Sent()[0]
			assert.Equal(t, ev.Type(), "com.amazon.sqs.message")
			assert.Equal(t, "arn:aws:sqs:us-fake-0:123456789012:MyQueue", ev.Source())
			assert.Equal(t, ev.Subject(), tSenderID)

			// final assertions
			assert.Len(t, ceClient.Sent(), tc.numMsgs, "Received more events than expected")
			assert.Equal(t, tc.expectDeletes, sqsClient.callsDelete, "Called delete function more times than expected")
			assert.Empty(t, sqsClient.inFlightMsgs, "Found unprocessed in-flight messages")
		})
	}
}

// makeARN returns a fake SQS ARN for the given resource.
func makeARN(resource string) arn.ARN {
	return arn.ARN{
		Partition: "aws",
		Service:   "sqs",
		Region:    "us-fake-0",
		AccountID: "123456789012",
		Resource:  resource,
	}
}

// standardMockSQSClient is a mocked SQS client which returns a standard set of
// responses and never errors.
type standardMockSQSClient struct {
	sqsiface.SQSAPI

	sync.Mutex
	availMsgs    []*sqs.Message
	inFlightMsgs []*sqs.Message

	callsDelete int
}

func (*standardMockSQSClient) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) { //nolint:golint,stylecheck
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String(tQueueURL),
	}, nil
}

func (c *standardMockSQSClient) ReceiveMessageWithContext(_ context.Context,
	in *sqs.ReceiveMessageInput, _ ...request.Option) (*sqs.ReceiveMessageOutput, error) {

	c.Lock()
	defer c.Unlock()

	n := int(*in.MaxNumberOfMessages)
	if l := len(c.availMsgs); l < n {
		n = l
	}

	msgs := c.availMsgs[:n]

	c.availMsgs = c.availMsgs[n:]
	c.inFlightMsgs = append(c.inFlightMsgs, msgs...)

	return &sqs.ReceiveMessageOutput{
		Messages: msgs,
	}, nil
}

func (c *standardMockSQSClient) DeleteMessageBatchWithContext(_ context.Context,
	in *sqs.DeleteMessageBatchInput, _ ...request.Option) (*sqs.DeleteMessageBatchOutput, error) {

	c.Lock()
	defer c.Unlock()

	c.callsDelete++

	inFlightIdx := make(map[ /*msg ID*/ string]int, len(c.inFlightMsgs))
	for i, msg := range c.inFlightMsgs {
		inFlightIdx[*msg.MessageId] = i
	}

	// mark processed messages by setting them to nil
	for _, msg := range in.Entries {
		if idx, ok := inFlightIdx[*msg.Id]; ok {
			c.inFlightMsgs[idx] = nil
		}
	}

	// filter nil entries in place
	oldInFlightMsgs := c.inFlightMsgs
	c.inFlightMsgs = c.inFlightMsgs[:0]
	for _, msg := range oldInFlightMsgs {
		if msg != nil {
			c.inFlightMsgs = append(c.inFlightMsgs, msg)
		}
	}

	return &sqs.DeleteMessageBatchOutput{}, nil
}

// makeMockMessages returns a set of mocked Messages.
func makeMockMessages(n int) []*sqs.Message {
	const receiptHandle = "dHJpZ2dlcm1lc2g="

	msgs := make([]*sqs.Message, n)

	for i := 0; i < n; i++ {
		msgs[i] = &sqs.Message{
			MessageId:     aws.String(fmt.Sprintf(tMsgIDPrefix+"%03d", i+1)),
			ReceiptHandle: aws.String(receiptHandle),
			Attributes: aws.StringMap(map[string]string{
				sqs.MessageSystemAttributeNameSenderId: tSenderID,
			}),
		}
	}

	return msgs
}

// Test that our mock implementation does what we expect.
func TestReceiveMessageWithContext(t *testing.T) {
	const rcvMsgs = 3
	const availMsgs = 4

	sqsClient := &standardMockSQSClient{
		availMsgs: makeMockMessages(availMsgs),
	}

	in := &sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(rcvMsgs),
	}

	expectRcv := availMsgs - rcvMsgs

	expectInFlight := []*sqs.Message{
		sqsClient.availMsgs[0],
		sqsClient.availMsgs[1],
		sqsClient.availMsgs[2],
	}

	_, err := sqsClient.ReceiveMessageWithContext(context.Background(), in)
	assert.NoError(t, err)

	assert.Len(t, sqsClient.availMsgs, expectRcv)
	assert.EqualValues(t, expectInFlight, sqsClient.inFlightMsgs)
}

// Test that our mock implementation does what we expect.
func TestDeleteMessageBatchWithContext(t *testing.T) {
	const inFlightMsgs = 5

	sqsClient := &standardMockSQSClient{
		inFlightMsgs: makeMockMessages(inFlightMsgs),
	}

	in := &sqs.DeleteMessageBatchInput{
		Entries: []*sqs.DeleteMessageBatchRequestEntry{{
			Id: sqsClient.inFlightMsgs[1].MessageId,
		}, {
			Id: sqsClient.inFlightMsgs[2].MessageId,
		}},
	}

	expect := []*sqs.Message{
		sqsClient.inFlightMsgs[0],
		sqsClient.inFlightMsgs[3],
		sqsClient.inFlightMsgs[4],
	}

	_, err := sqsClient.DeleteMessageBatchWithContext(context.Background(), in)
	assert.NoError(t, err)

	assert.EqualValues(t, expect, sqsClient.inFlightMsgs)
	assert.Equal(t, 1, sqsClient.callsDelete)
}
