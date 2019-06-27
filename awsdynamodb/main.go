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
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"

	log "github.com/sirupsen/logrus"
)

//DynamoDBEvent represents AWS Dynamo DB payload
type DynamoDBEvent struct {
	AwsRegion    *string                       `locationName:"awsRegion" type:"string"`
	Dynamodb     *dynamodbstreams.StreamRecord `locationName:"dynamodb" type:"structure"`
	EventID      *string                       `locationName:"eventID" type:"string"`
	EventName    *string                       `locationName:"eventName" type:"string" enum:"OperationType"`
	EventSource  *string                       `locationName:"eventSource" type:"string"`
	EventVersion *string                       `locationName:"eventVersion" type:"string"`
	UserIdentity *dynamodbstreams.Identity     `locationName:"userIdentity" type:"structure"`
}

//Clients struct represent all clients
type Clients struct {
	DynamoDBStream dynamodbstreamsiface.DynamoDBStreamsAPI
	CloudEvents    cloudevents.Client
}

var (
	sink                   string
	accountAccessKeyID     string
	accountSecretAccessKey string
	accountRegion          string
	tableName              string
)

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	accountRegion = os.Getenv("AWS_REGION")
	tableName = os.Getenv("TABLE")
}

func main() {

	flag.Parse()

	//Set logging output levels
	_, varPresent := os.LookupEnv("DEBUG")
	if varPresent {
		log.SetLevel(log.DebugLevel)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(accountRegion),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
	})

	if err != nil {
		log.Fatalf("NewSession failed: %v", err)
	}

	dynamoDBStream := dynamodbstreams.New(sess)

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
		DynamoDBStream: dynamoDBStream,
		CloudEvents:    c,
	}

	log.Info("Begin listening for aws dynamo db streams")

	streams, err := clients.getStreams()
	if err != nil {
		log.Error("getStreams failed ", err)
	}

	log.Debug("Streams: ", streams)

	streamsDescriptions, err := clients.getStreamsDescriptions(streams)
	if err != nil {
		log.Error("getStreamsDescriptions failed ", err)
	}

	log.Debug("Streams Descriptions: ", streamsDescriptions)

	for _, streamDescription := range streamsDescriptions {
		if *streamDescription.StreamStatus != "ENABLED" {
			log.Infof("Stream for table [%v] is not enabled!", *streamDescription.TableName)
		}
	}

	for {

		shardIterators, err := clients.getShardIterators(streamsDescriptions)
		if err != nil {
			log.Error("getShardIterators failed ", err)
		}
		var wg sync.WaitGroup
		wg.Add(len(shardIterators))
		for _, shardIterator := range shardIterators {
			go clients.processLatestRecords(shardIterator)
			wg.Done()
		}
		wg.Wait()
	}
}

func (clients Clients) getStreams() ([]*dynamodbstreams.Stream, error) {
	streams := []*dynamodbstreams.Stream{}

	listStreamsInput := dynamodbstreams.ListStreamsInput{
		TableName: aws.String(tableName),
	}

	for {
		listStreamOutput, err := clients.DynamoDBStream.ListStreams(&listStreamsInput)
		if err != nil {
			return streams, err
		}

		streams = append(streams, listStreamOutput.Streams...)

		listStreamsInput.ExclusiveStartStreamArn = listStreamOutput.LastEvaluatedStreamArn

		if listStreamOutput.LastEvaluatedStreamArn == nil {
			break
		}
	}

	return streams, nil
}

func (clients Clients) getStreamsDescriptions(streams []*dynamodbstreams.Stream) ([]*dynamodbstreams.StreamDescription, error) {
	streamsDescriptions := []*dynamodbstreams.StreamDescription{}

	for _, stream := range streams {

		describeStreamOutput, err := clients.DynamoDBStream.DescribeStream(&dynamodbstreams.DescribeStreamInput{
			StreamArn: stream.StreamArn,
		})

		if err != nil {
			return streamsDescriptions, err
		}

		streamsDescriptions = append(streamsDescriptions, describeStreamOutput.StreamDescription)
	}
	return streamsDescriptions, nil
}

func (clients Clients) getShardIterators(streamsDescriptions []*dynamodbstreams.StreamDescription) ([]*string, error) {
	shardIterators := []*string{}

	for _, streamDescription := range streamsDescriptions {
		for _, shard := range streamDescription.Shards {
			getShardIteratorInput := dynamodbstreams.GetShardIteratorInput{
				ShardId:           shard.ShardId,
				ShardIteratorType: aws.String("LATEST"),
				StreamArn:         streamDescription.StreamArn,
			}

			result, err := clients.DynamoDBStream.GetShardIterator(&getShardIteratorInput)
			if err != nil {
				return shardIterators, err
			}

			shardIterators = append(shardIterators, result.ShardIterator)
		}
	}

	return shardIterators, nil
}

func (clients Clients) processLatestRecords(shardIterator *string) error {

	getRecordsInput := dynamodbstreams.GetRecordsInput{
		ShardIterator: shardIterator,
	}

	getRecordsOutput, err := clients.DynamoDBStream.GetRecords(&getRecordsInput)
	if err != nil {
		log.Error("Get Records Failed: ", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(getRecordsOutput.Records))
	for _, record := range getRecordsOutput.Records {

		go func(record *dynamodbstreams.Record) {
			err := clients.sendDynamoDBEvent(record)
			if err != nil {
				log.Error("sendCloudEvent failed: ", err)
			}
		}(record)
		wg.Done()
	}

	wg.Wait()

	return nil
}

func (clients Clients) sendDynamoDBEvent(record *dynamodbstreams.Record) error {

	log.Info("Processing record ID: ", record.EventID)

	dynamoDBEvent := &DynamoDBEvent{
		AwsRegion:    record.AwsRegion,
		Dynamodb:     record.Dynamodb,
		EventID:      record.EventID,
		EventName:    record.EventName,
		EventSource:  record.EventSource,
		EventVersion: record.EventVersion,
		UserIdentity: record.UserIdentity,
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			Type:            "com.amazon.dynamodb",
			Subject:         aws.String("AWS Dynamo DB"),
			ID:              *record.EventID,
			Source:          *types.ParseURLRef(*record.EventSource),
			DataContentType: aws.String("application/json"),
		}.AsV03(),
		Data: dynamoDBEvent,
	}

	_, err := clients.CloudEvents.Send(context.Background(), event)
	if err != nil {
		return err
	}

	return nil
}
