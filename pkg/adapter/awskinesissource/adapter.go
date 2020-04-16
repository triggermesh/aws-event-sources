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

package awskinesissource

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	Stream                 string `envconfig:"STREAM" required:"true"`
	AWSRegion              string `envconfig:"AWS_REGION" required:"true"`
	AccountAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AccountSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	knsClient kinesisiface.KinesisAPI
	ceClient  cloudevents.Client

	stream                 string
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

	// create Kinesis client
	sess, err := session.NewSession(&aws.Config{
		Region:      &env.AWSRegion,
		Credentials: credentials.NewStaticCredentials(env.AccountAccessKeyId, env.AccountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logger.Fatalw("Failed to create Kinesis client", "error", err)
	}

	return &adapter{
		logger: logger,

		knsClient: kinesis.New(sess),
		ceClient:  ceClient,

		stream:                 env.Stream,
		awsRegion:              env.AWSRegion,
		accountAccessKeyID:     env.AccountAccessKeyId,
		accountSecretAccessKey: env.AccountSecretAccessKey,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	// Get info about a particular stream
	myStream, err := a.knsClient.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: &a.stream,
	})
	if err != nil {
		a.logger.Fatalw("Failed to describe stream", "error", err)
	}

	streamARN := myStream.StreamDescription.StreamARN

	a.logger.Infof("Connected to Kinesis stream: %s", *streamARN)

	// Obtain records inputs for different shards
	inputs, shardIDs := a.getRecordsInputs(myStream.StreamDescription.Shards)

	for {
		err := a.processInputs(inputs, shardIDs, streamARN)
		if err != nil {
			a.logger.Errorw("Failed to process inputs", "error", err)
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
			a.logger.Errorw("Failed to get shard iterator", "error", err)
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
	for i, input := range inputs {
		shardID := shardIDs[i]

		recordsOutput, err := a.knsClient.GetRecords(&input)
		if err != nil {
			a.logger.Errorw("Failed to get records", "error", err)
			continue
		}

		for _, record := range recordsOutput.Records {
			err := a.sendKinesisRecord(record, shardID, streamARN)
			if err != nil {
				a.logger.Errorw("Failed to send CloudEvent", "error", err)
			}
		}

		// remove old imput
		inputs = append(inputs[:i], inputs[i+1:]...)

		// generate new input
		input = kinesis.GetRecordsInput{
			ShardIterator: recordsOutput.NextShardIterator,
		}

		// add newly generated input to the slice
		// so that new iteration would begin with new sharIterator
		inputs = append(inputs, input)
	}

	return nil
}

func (a *adapter) sendKinesisRecord(record *kinesis.Record, shardID, streamARN *string) error {
	a.logger.Infof("Processing record ID: %s", *record.SequenceNumber)

	data := &Event{
		EventName:      aws.String("aws:kinesis:record"),
		EventSourceARN: streamARN,
		EventSource:    aws.String("aws:kinesis"),
		AWSRegion:      &a.awsRegion,
		Kinesis: Kinesis{
			PartitionKey:         record.PartitionKey,
			Data:                 record.Data,
			SequenceNumber:       record.SequenceNumber,
			KinesisSchemaVersion: aws.String("1.0"),
		},
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSKinesisEventType(v1alpha1.AWSKinesisGenericEventType))
	event.SetSubject(a.stream)
	event.SetSource(v1alpha1.AWSKinesisEventSource(a.awsRegion, a.stream))
	event.SetID(fmt.Sprintf("%s:%s", *shardID, *record.SequenceNumber))
	event.SetData(cloudevents.ApplicationJSON, data)

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
