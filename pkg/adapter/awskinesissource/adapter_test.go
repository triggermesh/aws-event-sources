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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/stretchr/testify/assert"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
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
	streamARN := aws.String("testARN")

	now := time.Now()
	records := []*kinesis.Record{
		{
			SequenceNumber:              aws.String("1"),
			PartitionKey:                aws.String("key"),
			ApproximateArrivalTimestamp: &now,
			Data:                        []byte("foo"),
		},
	}

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: adaptertest.NewTestClient(),
	}

	a.knsClient = mockedGetRecords{
		Resp: kinesis.GetRecordsOutput{
			NextShardIterator: aws.String("nextIterator"),
			Records:           records,
		},
		err: nil,
	}

	inputs := []kinesis.GetRecordsInput{
		{},
	}

	err := a.processInputs(inputs, []*string{aws.String("shardID")}, streamARN)
	assert.NoError(t, err)

	a.knsClient = mockedGetRecords{
		Resp: kinesis.GetRecordsOutput{},
		err:  errors.New("fake error"),
	}

	err = a.processInputs(inputs, []*string{aws.String("shardID")}, streamARN)
	assert.NoError(t, err)
}

func TestGetRecordsInputs(t *testing.T) {
	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	a.knsClient = mockedGetShardIterator{
		Resp: kinesis.GetShardIteratorOutput{ShardIterator: aws.String("shardIterator")},
		err:  nil,
	}

	shards := []*kinesis.Shard{
		{ShardId: aws.String("1")},
	}

	inputs, _ := a.getRecordsInputs(shards)
	assert.Equal(t, 1, len(inputs))

	a.knsClient = mockedGetShardIterator{
		Resp: kinesis.GetShardIteratorOutput{},
		err:  errors.New("fake error"),
	}

	inputs, _ = a.getRecordsInputs(shards)
	assert.Equal(t, 0, len(inputs))
}

func TestSendCloudevent(t *testing.T) {
	streamARN := aws.String("testARN")

	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		stream:   "fooStream",
		ceClient: ceClient,
	}

	record := kinesis.Record{
		Data:           []byte("foo"),
		SequenceNumber: aws.String("1"),
		PartitionKey:   aws.String("key"),
	}

	err := a.sendKinesisRecord(&record, aws.String(""), streamARN)
	assert.NoError(t, err)

	gotEvents := ceClient.Sent()
	assert.Len(t, gotEvents, 1, "Expected 1 event, got %d", len(gotEvents))

	wantData := `{"eventID":null,"eventVersion":null,"kinesis":{"partitionKey":"key","data":"Zm9v","sequenceNumber":"1","kinesisSchemaVersion":"1.0"},"eventName":"aws:kinesis:record","eventSourceARN":"testARN","eventSource":"aws:kinesis","awsRegion":""}`
	gotData := string(gotEvents[0].Data())
	assert.EqualValues(t, wantData, gotData, "Expected event %q, got %q", wantData, gotData)
}
