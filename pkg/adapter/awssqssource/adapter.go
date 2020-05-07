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

const waitTimeoutSec = 20

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	url, err := a.queueLookup(a.arn.Resource)
	if err != nil {
		a.logger.Fatalw("Unable to find URL of SQS queue "+a.arn.Resource, zap.Error(err))
	}

	queueURL := *url.QueueUrl
	a.logger.Infof("Listening to SQS queue at URL: %s", queueURL)

	// Look for new messages every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		msgs, err := a.getMessages(queueURL, waitTimeoutSec)
		if err != nil {
			a.logger.Errorw("Failed to get messages from SQS queue", "error", err)
			continue
		}

		// Only push if there are messages on the queue
		if len(msgs) < 1 {
			continue
		}

		attributes, err := a.sqsClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String("QueueArn")},
			QueueUrl:       url.QueueUrl,
		})
		if err != nil {
			a.logger.Errorw("Failed to get queue attributes", "error", err)
			continue
		}

		err = a.sendSQSEvent(msgs[0], attributes.Attributes["QueueArn"])
		if err != nil {
			a.logger.Errorw("Failed to send event", "error", err)
			continue
		}

		// Delete message from queue if we pushed successfully
		err = a.deleteMessage(queueURL, msgs[0].ReceiptHandle)
		if err != nil {
			a.logger.Errorw("Failed to delete message from SQS queue", "error", err)
			continue
		}
	}

	return nil
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
func (a *adapter) getMessages(queueURL string, waitTimeout int64) ([]*sqs.Message, error) {
	params := sqs.ReceiveMessageInput{
		AttributeNames: aws.StringSlice([]string{"All"}),
		QueueUrl:       &queueURL,
	}
	if waitTimeout > 0 {
		params.WaitTimeSeconds = &waitTimeout
	}
	resp, err := a.sqsClient.ReceiveMessage(&params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (a *adapter) sendSQSEvent(msg *sqs.Message, queueARN *string) error {
	a.logger.Infof("Processing message with ID: %s", *msg.MessageId)

	data := &Event{
		MessageID:         msg.MessageId,
		ReceiptHandle:     msg.ReceiptHandle,
		Body:              msg.Body,
		Attributes:        msg.Attributes,
		MessageAttributes: msg.MessageAttributes,
		Md5OfBody:         msg.MD5OfBody,
		EventSource:       aws.String("aws:sqs"),
		EventSourceARN:    queueARN,
		AwsRegion:         &a.arn.Region,
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSSQSGenericEventType))
	event.SetSubject(*msg.MessageId)
	event.SetSource(a.arn.String())
	event.SetID(*msg.MessageId)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
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
