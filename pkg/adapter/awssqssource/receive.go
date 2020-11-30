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
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

const (
	// Highest possible value for the MaxNumberOfMessages request parameter.
	// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
	maxReceiveMsgBatchSize = 10

	// Longest possible duration of a long polling request.
	// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html#sqs-long-polling
	maxLongPollingWaitTimeSeconds = 20

	// Visibility timeout to set on all messages received by this event source.
	// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html
	visibilityTimeoutSeconds = 30

	// Duration between calls to ReceiveMessage when the previous call didn't return any message.
	receiveMsgPeriod = 3 * time.Second
)

// A message receiver establishes long-lived connection to the SQS queue to
// fetch new messages.
func (a *adapter) runMessagesReceiver(ctx context.Context, queueURL string) {
	t := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return

		case <-t.C:
			messages, err := receiveMessages(ctx, a.sqsClient, queueURL)
			if err != nil {
				a.logger.Errorw("Failed to get messages from the SQS queue", zap.Error(err))
				t.Reset(1 * time.Second)
				continue
			}

			nextRequestDelay := receiveMsgPeriod
			if l := len(messages); l > 0 {
				// keep iterating immediately if any message was
				// received, so that bursts of new messages are
				// processed quickly
				nextRequestDelay = 0

				a.logger.Debugw("Received "+strconv.Itoa(l)+" message(s)",
					zap.Array(logfieldMsgID, messageList(messages)))
			}

			for _, msg := range messages {
				a.processQueue <- msg
			}

			t.Reset(nextRequestDelay)
		}
	}
}

type messageList []*sqs.Message

var _ zapcore.ArrayMarshaler = (messageList)(nil)

// MarshalLogArray implements zapcore.ArrayMarshaler.
func (ml messageList) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, m := range ml {
		arr.AppendString(*m.MessageId)
	}
	return nil
}

// receiveMessages returns a batch of messages read from the SQS queue, if any
// is available.
func receiveMessages(ctx context.Context, cli sqsiface.SQSAPI, queueURL string) ([]*sqs.Message, error) {
	resp, err := cli.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		AttributeNames:      aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: aws.Int64(maxReceiveMsgBatchSize),
		WaitTimeSeconds:     aws.Int64(maxLongPollingWaitTimeSeconds),
		VisibilityTimeout:   aws.Int64(visibilityTimeoutSeconds),
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}
