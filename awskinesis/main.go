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
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	cloudevents "github.com/cloudevents/sdk-go"
	log "github.com/sirupsen/logrus"
)

var (
	accountAccessKeyID     string
	accountSecretAccessKey string
	streamName             string
	region                 string
	sink                   string
	streamARN              *string
)

// Clients provides the ability to operate on Kinesis stream.
type Clients struct {
	Kinesis     kinesisiface.KinesisAPI
	CloudEvents cloudevents.Client
}

//Kinesis represents Kinesis stream item structure
type Kinesis struct {
	ParticionKey         *string `json:"partitionKey"`
	Data                 []byte  `json:"data"`
	SequenceNumber       *string `json:"sequenceNumber"`
	KinesisSchemaVersion *string `json:"kinesisSchemaVersion"`
}

// Event represents Amazon Kinesis Data Streams Event
type Event struct {
	EventID      *string `json:"eventID"`
	EventVersion *string `json:"eventVersion"`
	Kinesis      Kinesis `json:"kinesis"`
	//InvokeIdentityArn string `json:"invokeIdentityArn"`
	EventName      *string `json:"eventName"`
	EventSourceARN *string `json:"eventSourceARN"`
	EventSource    *string `json:"eventSource"`
	AWSRegion      *string `json:"awsRegion"`
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {
	flag.Parse()

	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = os.Getenv("AWS_REGION")
	streamName = os.Getenv("STREAM")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}

	t, err := cloudevents.NewHTTPTransport()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow())
	if err != nil {
		log.Fatal(err)
	}

	clients := Clients{kinesis.New(sess), c}

	// Get info about a particular stream
	myStream, err := clients.Kinesis.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})
	if err != nil {
		log.Fatal(err)
	}

	streamARN = myStream.StreamDescription.StreamARN

	//Obtain records inputs for different shards
	inputs, shardIDs := clients.getRecordsInputs(myStream.StreamDescription.Shards)

	log.Info("Connected to Kinesis stream: ", streamARN)

	for {
		err := clients.processInputs(inputs, shardIDs)
		if err != nil {
			log.Error(err)
		}
	}
}

func (clients Clients) getRecordsInputs(shards []*kinesis.Shard) ([]kinesis.GetRecordsInput, []*string) {
	inputs := []kinesis.GetRecordsInput{}
	shardIDs := []*string{}

	//Kinesis stream might have several shards and each of them had "LATEST" Iterator.
	for _, shard := range shards {

		// Obtain starting Shard Iterator. This is needed to not process already processed records
		myShardIterator, err := clients.Kinesis.GetShardIterator(&kinesis.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("LATEST"),
			StreamName:        aws.String(streamName),
		})

		if err != nil {
			log.Error(err)
			continue
		}

		// set records output limit. Should not be more than 10000, othervise panics
		input := kinesis.GetRecordsInput{
			ShardIterator: myShardIterator.ShardIterator,
		}

		inputs = append(inputs, input)
		shardIDs = append(shardIDs, shard.ShardId)
	}

	return inputs, shardIDs
}

func (clients Clients) processInputs(inputs []kinesis.GetRecordsInput, shardIDs []*string) error {

	for i, input := range inputs {
		shardID := shardIDs[i]

		recordsOutput, err := clients.Kinesis.GetRecords(&input)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, record := range recordsOutput.Records {
			err := clients.sendCognitoEvent(record, shardID)
			if err != nil {
				log.Errorf("SendCloudEvent failed: %v", err)
			}
		}

		//remove old imput
		inputs = append(inputs[:i], inputs[i+1:]...)

		//generate new input
		input = kinesis.GetRecordsInput{
			ShardIterator: recordsOutput.NextShardIterator,
		}

		//add newly generated input to the slice
		//so that new iteration would begin with new sharIterator
		inputs = append(inputs, input)
	}

	return nil
}

func (clients Clients) sendCognitoEvent(record *kinesis.Record, shardID *string) error {
	log.Info("Processing record ID: ", *record.SequenceNumber)

	kinesisEvent := &Event{
		EventName:      aws.String("aws:kinesis:record"),
		EventSourceARN: streamARN,
		EventSource:    aws.String("aws:kinesis"),
		AWSRegion:      aws.String(region),
		Kinesis: Kinesis{
			ParticionKey:         record.PartitionKey,
			Data:                 record.Data,
			SequenceNumber:       record.SequenceNumber,
			KinesisSchemaVersion: aws.String("1.0"),
		},
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			Type:            "AWS Kinesis Record",
			Subject:         aws.String("AWS Kinesis"),
			ID:              fmt.Sprintf("%s:%s", *shardID, *record.SequenceNumber),
			SpecVersion:     "1.0",
			DataContentType: aws.String("application/json"),
		}.AsV03(),
		Data: kinesisEvent,
	}

	_, err := clients.CloudEvents.Send(context.Background(), event)
	if err != nil {
		return err
	}

	return nil
}
