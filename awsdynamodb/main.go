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

	streams := []*dynamodbstreams.Stream{}

	listStreamsInput := dynamodbstreams.ListStreamsInput{
		TableName: aws.String(tableName),
	}

	for {
		listStreamOutput, err := client.StreamsClient.ListStreams(&listStreamsInput)
		if err != nil {
			log.Error(err)
		}

		streams = append(streams, listStreamOutput.Streams...)

		if listStreamOutput.LastEvaluatedStreamArn == nil {
			break
		}
		listStreamsInput.ExclusiveStartStreamArn = listStreamOutput.LastEvaluatedStreamArn
	}

	streamDescriptions := []*dynamodbstreams.StreamDescription{}

	for _, stream := range streams {

		describeStreamOutput, err := client.StreamsClient.DescribeStream(&dynamodbstreams.DescribeStreamInput{
			StreamArn: stream.StreamArn,
		})

		if err != nil {
			log.Error(err)
		}

		streamDescriptions = append(streamDescriptions, describeStreamOutput.StreamDescription)
	}

	shardIterators := []*string{}

	for _, streamDescription := range streamDescriptions {
		for _, shard := range streamDescription.Shards {
			getShardIteratorInput := dynamodbstreams.GetShardIteratorInput{
				ShardId:           shard.ShardId,
				ShardIteratorType: aws.String("LATEST"),
				StreamArn:         streamDescription.StreamArn,
			}

			_, result := client.StreamsClient.GetShardIteratorRequest(&getShardIteratorInput)

			shardIterators = append(shardIterators, result.ShardIterator)
		}
	}

	// Get Records out of shard iterators.
	records := []*dynamodbstreams.Record{}

	for _, shardIterator := range shardIterators {
		getRecordsInput := dynamodbstreams.GetRecordsInput{
			ShardIterator: shardIterator,
		}

		for {

			getRecordsOutput, err := client.StreamsClient.GetRecords(&getRecordsInput)
			if err != nil {
				log.Error(err)
			}

			records = append(records, getRecordsOutput.Records...)

			if getRecordsOutput.NextShardIterator == nil {
				break
			}
			getRecordsInput.ShardIterator = getRecordsOutput.NextShardIterator

		}
	}

	for _, record := range records {
		err := client.sendCloudevent(record)
		if err != nil {
			log.Error(err)
		}
	}

}

func (c Client) sendCloudevent(record *dynamodbstreams.Record) error {

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
