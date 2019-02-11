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
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/knative/pkg/cloudevents"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	log "github.com/sirupsen/logrus"
)

var sink string
var region string

// Queue provides the ability to handle SQS messages.
type Queue struct {
	Client sqsiface.SQSAPI
	URL    *string
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

	q := Queue{
		Client: sqsClient,
	}

	queueURL, err := q.QueueLookup(queueName)
	if err != nil {
		log.Fatal("Unable to find queue. Error: ", err)
	}

	log.Info("Beginning to listen at URL: ", queueURL)

	q.URL = aws.String(queueURL)

	//Look for new messages every 5 seconds
	for range time.Tick(5 * time.Second) {
		msgs, err := q.GetMessages(20)
		if err != nil {
			log.Error(err)
			continue
		}

		//Only push if there are messages on the queue
		if len(msgs) < 1 {
			continue
		}

		attributes, err := q.Client.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String("QueueArn")},
			QueueUrl:       q.URL,
		})

		c := cloudevents.NewClient(
			sink,
			cloudevents.Builder{
				Source:    "aws:sqs",
				EventType: "SQS message",
			},
		)

		err = pushMessage(c, msgs[0], attributes.Attributes["QueueArn"])
		if err != nil {
			log.Error(err)
			continue
		}

		//Delete message from queue if we pushed successfully
		err = q.DeleteMessage(msgs[0].ReceiptHandle)
		if err != nil {
			log.Error(err)
			continue
		}

	}

}

//QueueLookup finds the URL for a given queue name in the user's env.
//Needs to be an exact match to queue name and queue must be unique name in the AWS account.
func (q *Queue) QueueLookup(queueName string) (string, error) {
	queues, err := q.Client.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: aws.String(queueName)})
	if err != nil {
		return "", err
	}
	return aws.StringValue(queues.QueueUrls[0]), nil
}

// GetMessages returns the parsed messages from SQS if any. If an error
// occurs that error will be returned.
func (q *Queue) GetMessages(waitTimeout int64) ([]*sqs.Message, error) {
	params := sqs.ReceiveMessageInput{
		AttributeNames: aws.StringSlice([]string{"All"}),
		QueueUrl:       q.URL,
	}
	if waitTimeout > 0 {
		params.WaitTimeSeconds = aws.Int64(waitTimeout)
	}
	resp, err := q.Client.ReceiveMessage(&params)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func pushMessage(c *cloudevents.Client, msg *sqs.Message, queueARN *string) error {
	log.Info("Processing message with ID: ", aws.StringValue(msg.MessageId))
	log.Info(msg)

	sqsEvent := Event{
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

	if err := c.Send(sqsEvent); err != nil {
		log.Printf("error sending: %v", err)
	}

	return nil
}

//DeleteMessage deletes message from sqs queue
func (q *Queue) DeleteMessage(msg *string) error {
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      q.URL,
		ReceiptHandle: msg,
	}
	_, err := q.Client.DeleteMessage(deleteParams)
	if err != nil {
		return err
	}
	log.Info("Message deleted!")
	return nil
}
