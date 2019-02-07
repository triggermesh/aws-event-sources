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
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/triggermesh/sources/tmevents"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	log "github.com/sirupsen/logrus"
)

var (
	sink                   string
	accountAccessKeyID     string
	accountSecretAccessKey string
	queueEnv               string
	awsRegionEnv           string
)

type sqsMsg struct{}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	//TODO: Make sure all these env vars exist
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	queueEnv = os.Getenv("QUEUE")
	awsRegionEnv = os.Getenv("AWS_REGION")

	//Create client for SQS and start polling for messages on the queue
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegionEnv),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}
	sqsClient := sqs.New(sess)
	m := sqsMsg{}
	m.ReceiveMsg(sqsClient, sink)

}

//queueLookup finds the URL for a given queue name in the user's env.
//Needs to be an exact match to queue name and queue must be unique name in the AWS account.
func queueLookup(sqsClient *sqs.SQS) (string, error) {
	queues, err := sqsClient.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: &queueEnv})
	if err != nil {
		return "", err
	}
	return aws.StringValue(queues.QueueUrls[0]), nil

}

//ReceiveMsg implements the receive interface for sqs
func (sqsMsg) ReceiveMsg(sqsClient *sqs.SQS, sink string) {

	queueURL, err := queueLookup(sqsClient)
	if err != nil {
		log.Fatal("Unable to find queue. Error: ", err)
	}
	log.Info("Beginning to listen at URL: ", queueURL)

	//Look for new messages every 5 seconds
	for range time.Tick(5 * time.Second) {
		msg, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: aws.StringSlice([]string{"All"}),
			QueueUrl:       &queueURL,
		})
		if err != nil {
			log.Info(err)
			continue
		}
		//Only push if there are messages on the queue
		if len(msg.Messages) > 0 {
			log.Info("Processing message with ID: ", aws.StringValue(msg.Messages[0].MessageId))
			log.Info(msg.Messages[0])
			//Parse out timestamp
			msgAttributes := aws.StringValueMap(msg.Messages[0].Attributes)
			timeInt, err := strconv.ParseInt(msgAttributes["SentTimestamp"], 10, 64)
			if err != nil {
				log.Info(err)
				continue
			}
			timeSent := time.Unix(timeInt, 0)

			//Craft event info and push it
			eventInfo := tmevents.EventInfo{
				EventData:   []byte(aws.StringValue(msg.Messages[0].Body)),
				EventID:     aws.StringValue(msg.Messages[0].MessageId),
				EventTime:   timeSent,
				EventType:   "cloudevent.greet.you",
				EventSource: "sqs",
			}
			err = tmevents.PushEvent(&eventInfo, sink)
			if err != nil {
				log.Error(err)
				continue
			}

			//Delete message from queue if we pushed successfully
			err = deleteMessage(sqsClient, queueURL, msg)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}
}

//Deletes message from sqs queue
func deleteMessage(sqsClient *sqs.SQS, queueURL string, msg *sqs.ReceiveMessageOutput) error {
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.Messages[0].ReceiptHandle,
	}
	_, err := sqsClient.DeleteMessage(deleteParams)
	if err != nil {
		return err
	}
	return nil
}
