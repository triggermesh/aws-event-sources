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
	"time"

	"go.uber.org/zap"

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
			a.logger.Errorw("Failed to get messages from SQS queue", "error", err)
			return false, nil
		}

		for _, message := range messages {
			err = a.sendSQSEvent(message)
			if err != nil {
				a.logger.Errorw("Failed to send event", "error", err)
				continue
			}

			// Delete message from queue if we pushed successfully
			err = a.deleteMessage(queueURL, message.ReceiptHandle)
			if err != nil {
				a.logger.Errorw("Failed to delete message from SQS queue", "error", err)
				continue
			}
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
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:       &queueURL,
	}
	resp, err := a.sqsClient.ReceiveMessage(&params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (a *adapter) sendSQSEvent(msg *sqs.Message) error {
	a.logger.Infof("Processing message with ID: %s", *msg.MessageId)

	// TODO: work on CE attributes contract
	subject, exist := msg.Attributes[sqs.MessageSystemAttributeNameSenderId]
	if !exist {
		subject = msg.MessageId
	}
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSSQSGenericEventType))
	event.SetSubject(*subject)
	event.SetSource(a.arn.String())
	event.SetID(*msg.MessageId)
	if err := event.SetData(cloudevents.ApplicationJSON, msg); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

// deleteMessage deletes a message from the SQS queue.
func (a *adapter) deleteMessage(queueURL string, receiptHandle *string) error {
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: receiptHandle,
	}
	if _, err := a.sqsClient.DeleteMessage(deleteParams); err != nil {
		return err
	}

	a.logger.Debug("Message deleted")
	return nil
}
