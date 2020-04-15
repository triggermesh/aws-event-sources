/*
Copyright (c) 2020 TriggerMesh Inc.

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

package awsdynamodbsource

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	Table                  string `envconfig:"TABLE" required:"true"`
	AWSRegion              string `envconfig:"AWS_REGION" required:"true"`
	AccountAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AccountSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	dyndbClient dynamodbstreamsiface.DynamoDBStreamsAPI
	ceClient    cloudevents.Client

	table                  string
	awsRegion              string
	accountAccessKeyID     string
	accountSecretAccessKey string
}

func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor,
	ceClient cloudevents.Client) pkgadapter.Adapter {

	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	// create DynamoDB client
	sess, err := session.NewSession(&aws.Config{
		Region:      &env.AWSRegion,
		Credentials: credentials.NewStaticCredentials(env.AccountAccessKeyId, env.AccountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logger.Fatalw("Failed to create DynamoDB client", "error", err)
	}

	return &adapter{
		logger: logger,

		dyndbClient: dynamodbstreams.New(sess),
		ceClient:    ceClient,

		table:                  env.Table,
		awsRegion:              env.AWSRegion,
		accountAccessKeyID:     env.AccountAccessKeyId,
		accountSecretAccessKey: env.AccountSecretAccessKey,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	a.logger.Infof("Listening to AWS DynamoDB streams for table: %s", a.table)

	streams, err := a.getStreams()
	if err != nil {
		a.logger.Errorw("Failed to get Streams", "error", err)
	}

	a.logger.Debugf("Streams: %v", streams)

	streamsDescriptions, err := a.getStreamsDescriptions(streams)
	if err != nil {
		a.logger.Errorw("Failed to get Streams descriptions", "error", err)
	}

	a.logger.Debugf("Streams descriptions: %v", streamsDescriptions)

	for _, streamDescription := range streamsDescriptions {
		if *streamDescription.StreamStatus != "ENABLED" {
			a.logger.Infof("Stream for table %q is not enabled", *streamDescription.TableName)
		}
	}

	for {
		shardIterators, err := a.getShardIterators(streamsDescriptions)
		if err != nil {
			a.logger.Errorw("Failed to get shard iterators", "error", err)
		}

		var wg sync.WaitGroup
		wg.Add(len(shardIterators))

		for _, shardIterator := range shardIterators {
			go a.processLatestRecords(shardIterator)
			wg.Done()
		}
		wg.Wait()
	}
}

func (a *adapter) getStreams() ([]*dynamodbstreams.Stream, error) {
	streams := []*dynamodbstreams.Stream{}

	listStreamsInput := dynamodbstreams.ListStreamsInput{
		TableName: &a.table,
	}

	for {
		listStreamOutput, err := a.dyndbClient.ListStreams(&listStreamsInput)
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

func (a *adapter) getStreamsDescriptions(streams []*dynamodbstreams.Stream) ([]*dynamodbstreams.StreamDescription, error) {
	streamsDescriptions := []*dynamodbstreams.StreamDescription{}

	for _, stream := range streams {

		describeStreamOutput, err := a.dyndbClient.DescribeStream(&dynamodbstreams.DescribeStreamInput{
			StreamArn: stream.StreamArn,
		})

		if err != nil {
			return streamsDescriptions, err
		}

		streamsDescriptions = append(streamsDescriptions, describeStreamOutput.StreamDescription)
	}
	return streamsDescriptions, nil
}

func (a *adapter) getShardIterators(streamsDescriptions []*dynamodbstreams.StreamDescription) ([]*string, error) {
	shardIterators := []*string{}

	for _, streamDescription := range streamsDescriptions {
		for _, shard := range streamDescription.Shards {
			getShardIteratorInput := dynamodbstreams.GetShardIteratorInput{
				ShardId:           shard.ShardId,
				ShardIteratorType: aws.String("LATEST"),
				StreamArn:         streamDescription.StreamArn,
			}

			result, err := a.dyndbClient.GetShardIterator(&getShardIteratorInput)
			if err != nil {
				return shardIterators, err
			}

			shardIterators = append(shardIterators, result.ShardIterator)
		}
	}

	return shardIterators, nil
}

func (a *adapter) processLatestRecords(shardIterator *string) error {
	getRecordsInput := dynamodbstreams.GetRecordsInput{
		ShardIterator: shardIterator,
	}

	getRecordsOutput, err := a.dyndbClient.GetRecords(&getRecordsInput)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(getRecordsOutput.Records))

	for _, record := range getRecordsOutput.Records {
		go func(record *dynamodbstreams.Record) {
			err := a.sendDynamoDBEvent(record)
			if err != nil {
				a.logger.Errorw("Failed to send Cloud Event", "error", err)
			}
		}(record)
		wg.Done()
	}

	wg.Wait()

	return nil
}

func (a *adapter) sendDynamoDBEvent(record *dynamodbstreams.Record) error {
	a.logger.Infof("Processing record ID: %s", &record.EventID)

	data := &DynamoDBEvent{
		AwsRegion:    record.AwsRegion,
		Dynamodb:     record.Dynamodb,
		EventID:      record.EventID,
		EventName:    record.EventName,
		EventSource:  record.EventSource,
		EventVersion: record.EventVersion,
		UserIdentity: record.UserIdentity,
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSDynamoDBEventType(strings.ToLower(*record.EventName)))
	event.SetSubject(a.table)
	event.SetSource(v1alpha1.AWSDynamoDBEventSource(a.awsRegion, a.table))
	event.SetID(*record.EventID)
	event.SetData(cloudevents.ApplicationJSON, data)

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
