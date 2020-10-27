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
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// Highest possible value for the MaxNumberOfMessages request parameter.
// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
const maxReceiveMsgBatchSize = 10

// Highest possible duration of a long polling request.
// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html#sqs-long-polling
const maxLongPollingWaitTimeSeconds = 20

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN string `envconfig:"ARN" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	sqsClient sqsiface.SQSAPI
	ceClient  cloudevents.Client

	arn arn.ARN
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region).
		WithMaxRetries(5),
	))

	return &adapter{
		logger: logger,

		sqsClient: sqs.New(cfg),
		ceClient:  ceClient,

		arn: arn,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(ctx context.Context) error {
	url, err := a.queueLookup(a.arn.Resource)
	if err != nil {
		a.logger.Errorw("Unable to find URL of SQS queue "+a.arn.Resource, zap.Error(err))
		return err
	}

	queueURL := *url.QueueUrl
	a.logger.Infof("Listening to SQS queue at URL: %s", queueURL)

	backoff := common.NewBackoff(1 * time.Millisecond)

	err = backoff.Run(ctx.Done(), func(ctx context.Context) (bool, error) {
		messages, err := a.getMessages(queueURL)
		if err != nil {
			a.logger.Errorw("Failed to get messages from the SQS queue", zap.Error(err))
			return false, nil
		}

		if len(messages) == 0 {
			return true, nil
		}

		sent, err := a.sendSQSEvents(messages)
		if err != nil {
			// log the error but proceed with deleting the events
			// that were successfully sent to the sink
			a.logger.Errorw("Failed to send events to the sink", zap.Error(err))
		}

		err = a.deleteMessages(queueURL, sent)
		if err != nil {
			a.logger.Errorw("Failed to delete message from the SQS queue", zap.Error(err))
			return false, nil
		}

		return true, nil
	})

	return err
}

// queueLookup finds the URL for a given queue name in the user's env.
// Needs to be an exact match to queue name and queue must be unique name in the AWS account.
func (a *adapter) queueLookup(queueName string) (*sqs.GetQueueUrlOutput, error) {
	return a.sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
}

// getMessages returns the parsed messages from SQS if any. If an error
// occurs that error will be returned.
func (a *adapter) getMessages(queueURL string) ([]*sqs.Message, error) {
	params := sqs.ReceiveMessageInput{
		AttributeNames:      aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: aws.Int64(maxReceiveMsgBatchSize),
		WaitTimeSeconds:     aws.Int64(maxLongPollingWaitTimeSeconds),
	}
	resp, err := a.sqsClient.ReceiveMessage(&params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

// sendSQSEvents sends SQS messages to the event sink.
func (a *adapter) sendSQSEvents(msgs []*sqs.Message) (sent []*sqs.Message, err error) {
	// NOTE(antoineco): the CloudEvents SDK for Go doesn't support the
	// batched content mode, although it is defined in the spec.
	// https://github.com/cloudevents/spec/blob/v1.0/http-protocol-binding.md#33-batched-content-mode

	concurrentSent := concurrentMsgSlice{
		msgs: make([]*sqs.Message, 0, len(msgs)),
	}

	errCh := make(chan error, len(msgs))
	defer close(errCh)

	for _, msg := range msgs {
		a.logger.Debugf("Processing messages ID %s", *msg.MessageId)

		go func(msg *sqs.Message) {
			err := sendSQSEvent(a.ceClient, msg, a.arn)
			if err != nil {
				errCh <- fmt.Errorf("message ID %s: %w", *msg.MessageId, err)
				return
			}

			concurrentSent.append(msg)

			// always write to errCh to notify that sendSQSEvent returned
			errCh <- err
		}(msg)
	}

	var errs []error

	for i := 0; i < cap(errCh); i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		err = &errList{errs: errs}
	}

	return concurrentSent.msgs, err
}

// concurrentMsgSlice protects writes to a slice of Messages using a mutex.
type concurrentMsgSlice struct {
	sync.Mutex
	msgs []*sqs.Message
}

// append appends a new Message to the Messages slice in a thread-safe manner.
func (ms *concurrentMsgSlice) append(msg *sqs.Message) {
	ms.Lock()
	ms.msgs = append(ms.msgs, msg)
	ms.Unlock()
}

// sendSQSEvent sends a single SQS message to the event sink.
func sendSQSEvent(cli cloudevents.Client, msg *sqs.Message, queueARN arn.ARN) error {
	// TODO: work on CE attributes contract
	subject, exist := msg.Attributes[sqs.MessageSystemAttributeNameSenderId]
	if !exist {
		subject = msg.MessageId
	}

	event := cloudevents.NewEvent()
	event.SetType(v1alpha1.AWSEventType(queueARN.Service, v1alpha1.AWSSQSGenericEventType))
	event.SetSubject(*subject)
	event.SetSource(queueARN.String())
	event.SetID(*msg.MessageId)
	if err := event.SetData(cloudevents.ApplicationJSON, msg); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := cli.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

// deleteMessages deletes messages from the SQS queue.
func (a *adapter) deleteMessages(queueURL string, msgs []*sqs.Message) error {
	deleteEntries := make([]*sqs.DeleteMessageBatchRequestEntry, len(msgs))
	for i, msg := range msgs {
		deleteEntries[i] = &sqs.DeleteMessageBatchRequestEntry{
			Id:            msg.MessageId,
			ReceiptHandle: msg.ReceiptHandle,
		}
	}

	deleteParams := &sqs.DeleteMessageBatchInput{
		QueueUrl: &queueURL,
		Entries:  deleteEntries,
	}
	if _, err := a.sqsClient.DeleteMessageBatch(deleteParams); err != nil {
		return err
	}

	if a.logger.Desugar().Core().Enabled(zapcore.DebugLevel) {
		msgIds := make([]string, len(deleteEntries))
		for i, msg := range deleteEntries {
			msgIds[i] = *msg.Id
		}
		a.logger.Debugf("Deleted message IDs %v", msgIds)
	}

	return nil
}

type errList struct {
	errs []error
}

var _ error = (*errList)(nil)

// Error implements the error interface.
func (e *errList) Error() string {
	if e == nil || len(e.errs) == 0 {
		return ""
	}

	if len(e.errs) == 1 {
		return e.errs[0].Error()
	}

	var b strings.Builder

	b.WriteByte('[')
	for i, err := range e.errs {
		b.WriteString(err.Error())
		if i != len(e.errs)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteByte(']')

	return b.String()
}
