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
	"context"
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	log "github.com/sirupsen/logrus"
)

var sink string
var region string
var queueURL string

// Clients provides the ability to handle SQS messages.
type Clients struct {
	SQS         sqsiface.SQSAPI
	CloudEvents cloudevents.Client
}

// Event represent a sample of Amazon SQS Event
type Event struct {
	MessageID         *string                               `json:"messageId"`
	ReceiptHandle     *string                               `json:"receiptHandle"`
	Body              *string                               `json:"body"`
	Attributes        map[string]*string                    `json:"attributes"`
	MessageAttributes map[string]*sqs.MessageAttributeValue `json:"messageAttributes"`
	Md5OfBody         *string                               `json:"md5OfBody"`
	EventSource       *string                               `json:"eventSource"`
	EventSourceARN    *string                               `json:"eventSourceARN"`
	AwsRegion         *string                               `json:"awsRegion"`
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	//looks ugly, need to optimize
	accountAccessKeyID, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok {
		log.Fatal("AWS_ACCESS_KEY_ID env variable is not set!")
	}
	accountSecretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok {
		log.Fatal("AWS_SECRET_ACCESS_KEY env variable is not set!")
	}
	region, ok = os.LookupEnv("AWS_REGION")
	if !ok {
		log.Fatal("AWS_REGION env variable is not set!")
	}
	queueName, ok := os.LookupEnv("QUEUE")
	if !ok {
		log.Fatal("QUEUE env variable is not set!")
	}

	//Create client for SQS and start polling for messages on the queue
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}

	sqsClient := sqs.New(sess)
	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(sink),
	)
	if err != nil {
		log.Fatal(err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow())
	if err != nil {
		log.Fatal(err)
	}

	clients := Clients{
		SQS:         sqsClient,
		CloudEvents: c,
	}

	url, err := clients.QueueLookup(queueName)
	if err != nil {
		log.Fatal("Unable to find queue. Error: ", err)
	}

	log.Info("Beginning to listen at URL: ", url)

	queueURL = url

	//Look for new messages every 5 seconds
	for range time.Tick(5 * time.Second) {
		msgs, err := clients.GetMessages(20)
		if err != nil {
			log.Error(err)
			continue
		}

		//Only push if there are messages on the queue
		if len(msgs) < 1 {
			continue
		}

		attributes, err := clients.SQS.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String("QueueArn")},
			QueueUrl:       aws.String(queueURL),
		})

		err = clients.sendSQSEvent(msgs[0], attributes.Attributes["QueueArn"])
		if err != nil {
			log.Error(err)
			continue
		}

		//Delete message from queue if we pushed successfully
		err = clients.DeleteMessage(msgs[0].ReceiptHandle)
		if err != nil {
			log.Error(err)
			continue
		}

	}

}

//QueueLookup finds the URL for a given queue name in the user's env.
//Needs to be an exact match to queue name and queue must be unique name in the AWS account.
func (clients Clients) QueueLookup(queueName string) (string, error) {
	queues, err := clients.SQS.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: aws.String(queueName)})
	if err != nil {
		return "", err
	}
	return aws.StringValue(queues.QueueUrls[0]), nil
}

// GetMessages returns the parsed messages from SQS if any. If an error
// occurs that error will be returned.
func (clients Clients) GetMessages(waitTimeout int64) ([]*sqs.Message, error) {
	params := sqs.ReceiveMessageInput{
		AttributeNames: aws.StringSlice([]string{"All"}),
		QueueUrl:       aws.String(queueURL),
	}
	if waitTimeout > 0 {
		params.WaitTimeSeconds = aws.Int64(waitTimeout)
	}
	resp, err := clients.SQS.ReceiveMessage(&params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (clients Clients) sendSQSEvent(msg *sqs.Message, queueARN *string) error {
	log.Info("Processing message with ID: ", aws.StringValue(msg.MessageId))

	sqsEvent := &Event{
		MessageID:         msg.MessageId,
		ReceiptHandle:     msg.ReceiptHandle,
		Body:              msg.Body,
		Attributes:        msg.Attributes,
		MessageAttributes: msg.MessageAttributes,
		Md5OfBody:         msg.MD5OfBody,
		EventSource:       aws.String("aws:sqs"),
		EventSourceARN:    queueARN,
		AwsRegion:         aws.String(region),
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			Type:            "com.amazon.sqs.message",
			Source:          *types.ParseURLRef(queueURL),
			Subject:         aws.String("AWS SQS"),
			ID:              *msg.MessageId,
			DataContentType: aws.String("application/json"),
		}.AsV03(),
		Data: sqsEvent,
	}

	_, err := clients.CloudEvents.Send(context.Background(), event)
	return err
}

//DeleteMessage deletes message from sqs queue
func (clients Clients) DeleteMessage(msg *string) error {
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg,
	}
	_, err := clients.SQS.DeleteMessage(deleteParams)
	if err != nil {
		return err
	}
	log.Info("Message deleted!")
	return nil
}
