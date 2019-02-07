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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/triggermesh/sources/tmevents"
	"github.com/urakozz/go-dynamodb-stream-subscriber/stream"

	log "github.com/sirupsen/logrus"
)

var sink string

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	accountAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	accountSecureToken := os.Getenv("AWS_SECURE_TOKEN")
	accountRegion := os.Getenv("AWS_REGION")
	tableName := os.Getenv("TABLE")

	cfg := aws.NewConfig().WithRegion(accountRegion)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(accountRegion),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, accountSecureToken),
	})
	if err != nil {
		log.Errorf("NewSession failed: %v", err)
		return
	}
	streamSvc := dynamodbstreams.New(sess, cfg)
	dynamoSvc := dynamodb.New(sess, cfg)
	table := tableName

	log.Info("Begin listening for Dynamo DB Stream")

	streamSubscriber := stream.NewStreamSubscriber(dynamoSvc, streamSvc, table)
	ch, errCh := streamSubscriber.GetStreamDataAsync()

	go func(errCh <-chan error) {
		for err := range errCh {
			log.Errorf("Stream Subscriber error: %v", err)
		}
	}(errCh)

	for record := range ch {
		err := sendCloudevent(record, sink)
		if err != nil {
			log.Errorf("SendCloudEvent failed: %v", err)
		}
	}
}

func sendCloudevent(record *dynamodbstreams.Record, sink string) error {

	log.Info("Processing record ID: ", record.EventID)

	eventInfo := tmevents.EventInfo{
		EventData:   []byte(record.Dynamodb.String()),
		EventID:     *record.EventID,
		EventTime:   *record.Dynamodb.ApproximateCreationDateTime,
		EventType:   "cloudevent.greet.you",
		EventSource: *record.EventSource,
	}

	log.Infof("Pushing record [%v] to [%v] ", record.EventID, sink)

	err := tmevents.PushEvent(&eventInfo, sink)
	if err != nil {
		return err
	}
	return nil
}
