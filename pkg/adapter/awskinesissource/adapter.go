/*
Copyright (c) 2019-2020 TriggerMesh Inc.

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

package awskinesissource

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"

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

	knsClient kinesisiface.KinesisAPI
	ceClient  cloudevents.Client

	arn    arn.ARN
	stream string
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

		knsClient: kinesis.New(cfg),
		ceClient:  ceClient,

		arn:    arn,
		stream: common.MustParseKinesisResource(arn.Resource),
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	// Get info about a particular stream
	myStream, err := a.knsClient.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: &a.stream,
	})
	if err != nil {
		a.logger.Fatalw("Failed to describe stream", zap.Error(err))
	}

	streamARN := myStream.StreamDescription.StreamARN

	a.logger.Infof("Connected to Kinesis stream: %s", *streamARN)

	// Obtain records inputs for different shards
	inputs, shardIDs := a.getRecordsInputs(myStream.StreamDescription.Shards)

	for {
		err := a.processInputs(inputs, shardIDs, streamARN)
		if err != nil {
			a.logger.Errorw("Failed to process inputs", zap.Error(err))
		}
	}
}

func (a *adapter) getRecordsInputs(shards []*kinesis.Shard) ([]kinesis.GetRecordsInput, []*string) {
	inputs := []kinesis.GetRecordsInput{}
	shardIDs := []*string{}

	// Kinesis stream might have several shards and each of them had "LATEST" Iterator.
	for _, shard := range shards {
		// Obtain starting Shard Iterator. This is needed to not process already processed records
		myShardIterator, err := a.knsClient.GetShardIterator(&kinesis.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("LATEST"),
			StreamName:        &a.stream,
		})

		if err != nil {
			a.logger.Errorw("Failed to get shard iterator", zap.Error(err))
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

func (a *adapter) processInputs(inputs []kinesis.GetRecordsInput, shardIDs []*string, streamARN *string) error {
	var errs []error

	for i, input := range inputs {
		input := input

		recordsOutput, err := a.knsClient.GetRecords(&input)
		if err != nil {
			a.logger.Errorw("Failed to get records", zap.Error(err))
			errs = append(errs, err)
			continue
		}

		shardID := shardIDs[i]

		for _, record := range recordsOutput.Records {
			if err := a.sendKinesisRecord(record, shardID, streamARN); err != nil {
				a.logger.Errorw("Failed to send CloudEvent", zap.Error(err))
				errs = append(errs, err)
			}
		}

		// remove old input
		inputs = append(inputs[:i], inputs[i+1:]...)

		// generate new input
		input = kinesis.GetRecordsInput{
			ShardIterator: recordsOutput.NextShardIterator,
		}

		// add newly generated input to the slice
		// so that new iteration would begin with new sharIterator
		inputs = append(inputs, input)
	}

	return utilerrors.NewAggregate(errs)
}

func (a *adapter) sendKinesisRecord(record *kinesis.Record, shardID, streamARN *string) error {
	a.logger.Infof("Processing record ID: %s", *record.SequenceNumber)

	data := &Event{
		EventName:      aws.String("aws:kinesis:record"),
		EventSourceARN: streamARN,
		EventSource:    aws.String("aws:kinesis"),
		AWSRegion:      &a.arn.Region,
		Kinesis: Kinesis{
			PartitionKey:         record.PartitionKey,
			Data:                 record.Data,
			SequenceNumber:       record.SequenceNumber,
			KinesisSchemaVersion: aws.String("1.0"),
		},
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSKinesisGenericEventType))
	event.SetSubject(*record.PartitionKey)
	event.SetSource(a.arn.String())
	event.SetID(fmt.Sprintf("%s:%s", *shardID, *record.SequenceNumber))
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
