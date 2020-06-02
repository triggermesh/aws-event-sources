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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
)

type mockedDynamoStreamsClient struct {
	dynamodbstreamsiface.DynamoDBStreamsAPI
	listStreamsOutput      dynamodbstreams.ListStreamsOutput
	listStreamsOutputError error

	describeStreamOutput      dynamodbstreams.DescribeStreamOutput
	describeStreamOutputError error

	getShardIteratorOutput      dynamodbstreams.GetShardIteratorOutput
	getShardIteratorOutputError error

	getRecordsOutput      dynamodbstreams.GetRecordsOutput
	getRecordsOutputError error
}

func (m mockedDynamoStreamsClient) ListStreams(in *dynamodbstreams.ListStreamsInput) (*dynamodbstreams.ListStreamsOutput, error) {
	return &m.listStreamsOutput, m.listStreamsOutputError
}

func (m mockedDynamoStreamsClient) DescribeStream(in *dynamodbstreams.DescribeStreamInput) (*dynamodbstreams.DescribeStreamOutput, error) {
	return &m.describeStreamOutput, m.describeStreamOutputError
}

func (m mockedDynamoStreamsClient) GetShardIterator(in *dynamodbstreams.GetShardIteratorInput) (*dynamodbstreams.GetShardIteratorOutput, error) {
	return &m.getShardIteratorOutput, m.getShardIteratorOutputError
}

func (m mockedDynamoStreamsClient) GetRecords(in *dynamodbstreams.GetRecordsInput) (*dynamodbstreams.GetRecordsOutput, error) {
	return &m.getRecordsOutput, m.getRecordsOutputError
}

func TestGetStreams(t *testing.T) {
	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	a.dyndbClient = mockedDynamoStreamsClient{
		listStreamsOutput:      dynamodbstreams.ListStreamsOutput{},
		listStreamsOutputError: errors.New("fake getstreams error"),
	}

	streams, err := a.getStreams()
	assert.Error(t, err)
	assert.Equal(t, 0, len(streams))

	a.dyndbClient = mockedDynamoStreamsClient{
		listStreamsOutput: dynamodbstreams.ListStreamsOutput{
			Streams: []*dynamodbstreams.Stream{{}, {}},
		},
		listStreamsOutputError: nil,
	}

	streams, err = a.getStreams()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(streams))
}

func TestGetStreamsDescriptions(t *testing.T) {
	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	streams := []*dynamodbstreams.Stream{{}}

	a.dyndbClient = mockedDynamoStreamsClient{
		describeStreamOutput:      dynamodbstreams.DescribeStreamOutput{},
		describeStreamOutputError: errors.New("fake describestream error"),
	}

	streamsDescriptions, err := a.getStreamsDescriptions(streams)
	assert.Error(t, err)
	assert.Equal(t, 0, len(streamsDescriptions))

	streams = []*dynamodbstreams.Stream{
		{
			StreamArn:   aws.String("foo"),
			StreamLabel: aws.String("bar"),
			TableName:   aws.String("fooTable"),
		},
	}

	a.dyndbClient = mockedDynamoStreamsClient{
		describeStreamOutput: dynamodbstreams.DescribeStreamOutput{
			StreamDescription: &dynamodbstreams.StreamDescription{},
		},
		describeStreamOutputError: nil,
	}

	streamsDescriptions, err = a.getStreamsDescriptions(streams)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(streamsDescriptions))
}

func TestGetShardIterators(t *testing.T) {
	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	streamDescriptions := []*dynamodbstreams.StreamDescription{{
		Shards: []*dynamodbstreams.Shard{{}},
	}}

	a.dyndbClient = mockedDynamoStreamsClient{
		getShardIteratorOutput:      dynamodbstreams.GetShardIteratorOutput{},
		getShardIteratorOutputError: errors.New("fake getsharditerator error"),
	}

	streamsDescriptions, err := a.getShardIterators(streamDescriptions)
	assert.Error(t, err)
	assert.Equal(t, 0, len(streamsDescriptions))

	a.dyndbClient = mockedDynamoStreamsClient{
		describeStreamOutput: dynamodbstreams.DescribeStreamOutput{
			StreamDescription: &dynamodbstreams.StreamDescription{},
		},
		describeStreamOutputError: nil,
	}

	streamsDescriptions, err = a.getShardIterators(streamDescriptions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(streamsDescriptions))
}

func TestGetLatestRecords(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	shardIterator := aws.String("1")

	a.dyndbClient = mockedDynamoStreamsClient{
		getRecordsOutput:      dynamodbstreams.GetRecordsOutput{},
		getRecordsOutputError: errors.New("fake getrecords error"),
	}

	err := a.processLatestRecords(shardIterator)
	assert.Error(t, err)

	a.dyndbClient = mockedDynamoStreamsClient{
		getRecordsOutput: dynamodbstreams.GetRecordsOutput{
			Records: []*dynamodbstreams.Record{{
				EventID:     aws.String("1"),
				EventName:   aws.String("some event"),
				EventSource: aws.String("some source"),
			}},
		},
		getRecordsOutputError: nil,
	}

	err = a.processLatestRecords(shardIterator)
	assert.NoError(t, err)
}

func TestSendCloudevent(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		table:    "fooTable",
		ceClient: ceClient,
	}

	record := dynamodbstreams.Record{
		EventID:     aws.String("1"),
		EventName:   aws.String("some event"),
		EventSource: aws.String("some source"),
		Dynamodb: &dynamodbstreams.StreamRecord{
			Keys: map[string]*dynamodb.AttributeValue{"key1": nil, "key2": nil},
		},
	}

	// send multiple events to ensure we re-use buffers from strBuilderPool
	const sendEvents = 5

	for i := 0; i < sendEvents; i++ {
		err := a.sendDynamoDBEvent(&record)
		assert.NoError(t, err)
	}

	gotEvents := ceClient.Sent()
	assert.Len(t, gotEvents, sendEvents, "Expect %d sent events", sendEvents)

	wantData := `{"AwsRegion":null,"Dynamodb":{"ApproximateCreationDateTime":null,"Keys":{"key1":null,"key2":null},"NewImage":null,"OldImage":null,"SequenceNumber":null,"SizeBytes":null,"StreamViewType":null},"EventID":"1","EventName":"some event","EventSource":"some source","EventVersion":null,"UserIdentity":null}`

	for i := 0; i < sendEvents; i++ {
		gotData := string(gotEvents[i].Data())
		assert.EqualValues(t, wantData, gotData, "[%d] Compare sent data to expected", i)

		subject := gotEvents[i].Subject()
		assert.Contains(t, []string{"key1,key2", "key2,key1"}, subject,
			`[%d] Subject contains keys "key1,key2" in any order`, i)
	}
}
