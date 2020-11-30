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

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// A message processor processes SQS messages (sends as CloudEvent)
// sequentially, as soon as they are written to processQueue.
func (a *adapter) runMessagesProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-a.processQueue:
			a.logger.Debugw("Processing message", zap.String(logfieldMsgID, *msg.MessageId))

			if err := sendSQSEvent(ctx, a.ceClient, &a.arn, msg); err != nil {
				a.logger.Errorw("Failed to send event to the sink", zap.Error(err),
					zap.String(logfieldMsgID, *msg.MessageId))

				continue
			}

			a.deleteQueue <- msg
		}
	}
}

// sendSQSEvent sends a single SQS message as a CloudEvent to the event sink.
func sendSQSEvent(ctx context.Context, cli cloudevents.Client, arn *arn.ARN, msg *sqs.Message) error {
	// TODO: work on CE attributes contract
	subject, exist := msg.Attributes[sqs.MessageSystemAttributeNameSenderId]
	if !exist {
		subject = msg.MessageId
	}

	event := cloudevents.NewEvent()
	event.SetType(v1alpha1.AWSEventType(arn.Service, v1alpha1.AWSSQSGenericEventType))
	event.SetSubject(*subject)
	event.SetSource(arn.String())
	event.SetID(*msg.MessageId)
	if err := event.SetData(cloudevents.ApplicationJSON, msg); err != nil {
		return fmt.Errorf("setting CloudEvent data: %w", err)
	}

	if result := cli.Send(ctx, event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
