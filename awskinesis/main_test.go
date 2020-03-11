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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/cloudevents/sdk-go"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type mockedGetRecords struct {
	kinesisiface.KinesisAPI
	Resp kinesis.GetRecordsOutput
	err  error
}

type mockedGetShardIterator struct {
	kinesisiface.KinesisAPI
	Resp kinesis.GetShardIteratorOutput
	err  error
}

func (m mockedGetRecords) GetRecords(in *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
	return &m.Resp, m.err
}

func (m mockedGetShardIterator) GetShardIterator(in *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
	return &m.Resp, m.err
}

func TestProcessInputs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	now := time.Now()

	records := []*kinesis.Record{
		{
			SequenceNumber:              aws.String("1"),
			PartitionKey:                aws.String("key"),
			ApproximateArrivalTimestamp: &now,
			Data:                        []byte("foo"),
		},
	}

	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget("https://foo.com"),
	)
	assert.NoError(t, err)

	cloudClient, err := cloudevents.NewClient(transport)
	assert.NoError(t, err)

	clients := Clients{
		Kinesis: mockedGetRecords{Resp: kinesis.GetRecordsOutput{
			NextShardIterator: aws.String("nextIterator"),
			Records:           records,
		}, err: nil},
		CloudEvents: cloudClient,
	}

	inputs := []kinesis.GetRecordsInput{
		{},
	}

	err = clients.processInputs(inputs, []*string{aws.String("shardID")})
	assert.NoError(t, err)

	clients = Clients{
		Kinesis:     mockedGetRecords{Resp: kinesis.GetRecordsOutput{}, err: errors.New("error")},
		CloudEvents: cloudClient,
	}

	err = clients.processInputs(inputs, []*string{aws.String("shardID")})
	assert.NoError(t, err)
}

func TestGetRecordsInputs(t *testing.T) {

	clients := Clients{
		Kinesis: mockedGetShardIterator{Resp: kinesis.GetShardIteratorOutput{ShardIterator: aws.String("shardIterator")}, err: nil},
	}

	shards := []*kinesis.Shard{
		{ShardId: aws.String("1")},
	}

	inputs, _ := clients.getRecordsInputs(shards)
	assert.Equal(t, 1, len(inputs))

	clients = Clients{
		Kinesis: mockedGetShardIterator{Resp: kinesis.GetShardIteratorOutput{}, err: errors.New("err")},
	}

	inputs, _ = clients.getRecordsInputs(shards)
	assert.Equal(t, 0, len(inputs))

}

func TestSendCloudevent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	record := kinesis.Record{
		Data:           []byte("foo"),
		SequenceNumber: aws.String("1"),
		PartitionKey:   aws.String("key"),
	}

	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget("https://foo.com"),
	)
	assert.NoError(t, err)

	cloudClient, err := cloudevents.NewClient(transport)
	assert.NoError(t, err)

	clients := Clients{
		CloudEvents: cloudClient,
	}

	err = clients.sendKinesisRecord(&record, aws.String(""))
	assert.NoError(t, err)
}
