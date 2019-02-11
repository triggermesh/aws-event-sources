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
	"github.com/jarcoal/httpmock"
	"github.com/knative/pkg/cloudevents"
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
	sink = "https://foo.com"

	records := []*kinesis.Record{
		{
			SequenceNumber:              aws.String("1"),
			ApproximateArrivalTimestamp: &now,
			Data: []byte("foo"),
		},
	}

	c := cloudevents.NewClient(
		"https://foo.com",
		cloudevents.Builder{
			Source:    "aws:kinesis",
			EventType: "Kinesis Record",
		},
	)

	s := Stream{
		Client: mockedGetRecords{Resp: kinesis.GetRecordsOutput{
			NextShardIterator: aws.String("nextIterator"),
			Records:           records,
		}, err: nil},
	}

	s.cloudEventsClient = c

	inputs := []kinesis.GetRecordsInput{
		{},
	}

	err := s.processInputs(inputs, []*string{aws.String("shardID")})
	assert.NoError(t, err)

	s = Stream{
		Client: mockedGetRecords{Resp: kinesis.GetRecordsOutput{}, err: errors.New("error")},
	}

	s.cloudEventsClient = c

	err = s.processInputs(inputs, []*string{aws.String("shardID")})
	assert.NoError(t, err)

}

func TestGetRecordsInputs(t *testing.T) {
	s := Stream{
		Client: mockedGetShardIterator{Resp: kinesis.GetShardIteratorOutput{ShardIterator: aws.String("shardIterator")}, err: nil},
		Stream: aws.String("bar"),
	}

	shards := []*kinesis.Shard{
		{ShardId: aws.String("1")},
	}

	inputs, _ := s.getRecordsInputs(shards)
	assert.Equal(t, 1, len(inputs))

	s = Stream{
		Client: mockedGetShardIterator{Resp: kinesis.GetShardIteratorOutput{}, err: errors.New("err")},
		Stream: aws.String("bar"),
	}

	inputs, _ = s.getRecordsInputs(shards)
	assert.Equal(t, 0, len(inputs))

}

func TestSendCloudevent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	record := kinesis.Record{
		Data:           []byte("foo"),
		SequenceNumber: aws.String("1"),
	}

	c := cloudevents.NewClient(
		"https://foo.com",
		cloudevents.Builder{
			Source:    "aws:kinesis",
			EventType: "Kinesis Record",
		},
	)

	err := sendCloudevent(c, &record, aws.String(""))
	assert.NoError(t, err)
}
