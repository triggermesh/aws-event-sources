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

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN string `envconfig:"ARN" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	dyndbClient dynamodbstreamsiface.DynamoDBStreamsAPI
	ceClient    cloudevents.Client

	arn   arn.ARN
	table string
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region).
		WithMaxRetries(5),
	))

	return &adapter{
		logger: logger,

		dyndbClient: dynamodbstreams.New(cfg),
		ceClient:    ceClient,

		arn:   arn,
		table: common.MustParseDynamoDBResource(arn.Resource),
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	a.logger.Info("Listening to AWS DynamoDB streams for table: " + a.table)

	streams, err := a.getStreams()
	if err != nil {
		a.logger.Errorw("Failed to get Streams", zap.Error(err))
	}

	a.logger.Debugf("Streams: %v", streams)

	streamsDescriptions, err := a.getStreamsDescriptions(streams)
	if err != nil {
		a.logger.Errorw("Failed to get Streams descriptions", zap.Error(err))
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
			a.logger.Errorw("Failed to get shard iterators", zap.Error(err))
		}

		var wg sync.WaitGroup
		wg.Add(len(shardIterators))

		for _, shardIterator := range shardIterators {
			go func() {
				if err := a.processLatestRecords(shardIterator); err != nil {
					a.logger.Errorw("Error while processing records for shard iterator "+
						*shardIterator, zap.Error(err))
				}
			}()
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
			if err := a.sendDynamoDBEvent(record); err != nil {
				a.logger.Errorw("Failed to send CloudEvent", zap.Error(err))
			}
		}(record)
		wg.Done()
	}

	wg.Wait()

	return nil
}

func (a *adapter) sendDynamoDBEvent(record *dynamodbstreams.Record) error {
	a.logger.Info("Processing record ID: " + *record.EventID)

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
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, strings.ToLower(*record.EventName)))
	event.SetSubject(asEventSubject(record))
	event.SetSource(a.arn.String())
	event.SetID(*record.EventID)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

// asEventSubject returns an event subject corresponding to the given record.
func asEventSubject(record *dynamodbstreams.Record) string {
	if record == nil || record.Dynamodb == nil || record.Dynamodb.Keys == nil {
		return ""
	}

	subject := strBuilderPool.Get().(*strings.Builder)
	defer strBuilderPool.Put(subject)
	defer subject.Reset()

	i := 0
	for k := range record.Dynamodb.Keys {
		subject.WriteString(k)
		i++
		if i < len(record.Dynamodb.Keys) {
			subject.WriteByte(',')
		}
	}

	return subject.String()
}

var strBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}
