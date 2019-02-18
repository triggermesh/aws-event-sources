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
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"
	"github.com/knative/pkg/cloudevents"

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

//Client struct represent all clients
type Client struct {
	StreamsClient dynamodbstreamsiface.DynamoDBStreamsAPI
	CloudEvents   *cloudevents.Client
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

	streamsClient := dynamodbstreams.New(sess)

	cloudEvents := cloudevents.NewClient(
		sink,
		cloudevents.Builder{
			Source:    "aws:dynamodb",
			EventType: "dynamodb event",
		},
	)

	client := Client{
		StreamsClient: streamsClient,
		CloudEvents:   cloudEvents,
	}

	log.Info("Begin listening for aws dynamo db streams")

	streams, err := client.getStreams()
	if err != nil {
		log.Error("getStreams failed ", err)
	}

	log.Debug("Streams: ", streams)

	streamsDescriptions, err := client.getStreamsDescriptions(streams)
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

		shardIterators, err := client.getShardIterators(streamsDescriptions)
		if err != nil {
			log.Error("getShardIterators failed ", err)
		}
		var wg sync.WaitGroup
		wg.Add(len(shardIterators))
		for _, shardIterator := range shardIterators {
			go client.processLatestRecords(shardIterator)
			wg.Done()
		}
		wg.Wait()
	}
}

func (c Client) getStreams() ([]*dynamodbstreams.Stream, error) {
	streams := []*dynamodbstreams.Stream{}

	listStreamsInput := dynamodbstreams.ListStreamsInput{
		TableName: aws.String(tableName),
	}

	for {
		listStreamOutput, err := c.StreamsClient.ListStreams(&listStreamsInput)
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

func (c Client) getStreamsDescriptions(streams []*dynamodbstreams.Stream) ([]*dynamodbstreams.StreamDescription, error) {
	streamsDescriptions := []*dynamodbstreams.StreamDescription{}

	for _, stream := range streams {

		describeStreamOutput, err := c.StreamsClient.DescribeStream(&dynamodbstreams.DescribeStreamInput{
			StreamArn: stream.StreamArn,
		})

		if err != nil {
			return streamsDescriptions, err
		}

		streamsDescriptions = append(streamsDescriptions, describeStreamOutput.StreamDescription)
	}
	return streamsDescriptions, nil
}

func (c Client) getShardIterators(streamsDescriptions []*dynamodbstreams.StreamDescription) ([]*string, error) {
	shardIterators := []*string{}

	for _, streamDescription := range streamsDescriptions {
		for _, shard := range streamDescription.Shards {
			getShardIteratorInput := dynamodbstreams.GetShardIteratorInput{
				ShardId:           shard.ShardId,
				ShardIteratorType: aws.String("LATEST"),
				StreamArn:         streamDescription.StreamArn,
			}

			result, err := c.StreamsClient.GetShardIterator(&getShardIteratorInput)
			if err != nil {
				return shardIterators, err
			}

			shardIterators = append(shardIterators, result.ShardIterator)
		}
	}

	return shardIterators, nil
}

func (c Client) processLatestRecords(shardIterator *string) error {

	getRecordsInput := dynamodbstreams.GetRecordsInput{
		ShardIterator: shardIterator,
	}

	getRecordsOutput, err := c.StreamsClient.GetRecords(&getRecordsInput)
	if err != nil {
		log.Error("Get Records Failed: ", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(getRecordsOutput.Records))
	for _, record := range getRecordsOutput.Records {

		go func(record *dynamodbstreams.Record) {
			err := c.sendCloudEvent(record)
			if err != nil {
				log.Error("sendCloudEvent failed: ", err)
			}
		}(record)
		wg.Done()
	}

	wg.Wait()

	return nil
}

func (c Client) sendCloudEvent(record *dynamodbstreams.Record) error {

	log.Info("Processing record ID: ", record.EventID)

	dynamoDBEvent := DynamoDBEvent{
		AwsRegion:    record.AwsRegion,
		Dynamodb:     record.Dynamodb,
		EventID:      record.EventID,
		EventName:    record.EventName,
		EventSource:  record.EventSource,
		EventVersion: record.EventVersion,
		UserIdentity: record.UserIdentity,
	}

	if err := c.CloudEvents.Send(dynamoDBEvent); err != nil {
		return err
	}

	return nil
}
