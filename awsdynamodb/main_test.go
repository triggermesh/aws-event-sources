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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams/dynamodbstreamsiface"
	"github.com/jarcoal/httpmock"
	"github.com/knative/pkg/cloudevents"
	"github.com/stretchr/testify/assert"
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
	c := Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			listStreamsOutput:      dynamodbstreams.ListStreamsOutput{},
			listStreamsOutputError: errors.New("getstreams failed"),
		},
	}

	streams, err := c.getStreams()
	assert.Error(t, err)
	assert.Equal(t, 0, len(streams))

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			listStreamsOutput: dynamodbstreams.ListStreamsOutput{
				Streams: []*dynamodbstreams.Stream{{}, {}},
			},
			listStreamsOutputError: nil,
		},
	}

	streams, err = c.getStreams()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(streams))
}

func TestGetStreamsDescriptions(t *testing.T) {

	streams := []*dynamodbstreams.Stream{{}}

	c := Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			describeStreamOutput:      dynamodbstreams.DescribeStreamOutput{},
			describeStreamOutputError: errors.New("get stream description failed"),
		},
	}

	streamsDescriptions, err := c.getStreamsDescriptions(streams)
	assert.Error(t, err)
	assert.Equal(t, 0, len(streamsDescriptions))

	streams = []*dynamodbstreams.Stream{
		{
			StreamArn:   aws.String("foo"),
			StreamLabel: aws.String("bar"),
			TableName:   aws.String("fooTable"),
		},
	}

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			describeStreamOutput: dynamodbstreams.DescribeStreamOutput{
				StreamDescription: &dynamodbstreams.StreamDescription{},
			},
			describeStreamOutputError: nil,
		},
	}

	streamsDescriptions, err = c.getStreamsDescriptions(streams)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(streamsDescriptions))
}

func TestGetShardIterators(t *testing.T) {

	streamDescriptions := []*dynamodbstreams.StreamDescription{
		{
			Shards: []*dynamodbstreams.Shard{{}},
		},
	}

	c := Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			getShardIteratorOutput:      dynamodbstreams.GetShardIteratorOutput{},
			getShardIteratorOutputError: errors.New("failed to get stream description"),
		},
	}

	streamsDescriptions, err := c.getShardIterators(streamDescriptions)
	assert.Error(t, err)
	assert.Equal(t, 0, len(streamsDescriptions))

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			describeStreamOutput: dynamodbstreams.DescribeStreamOutput{
				StreamDescription: &dynamodbstreams.StreamDescription{},
			},
			describeStreamOutputError: nil,
		},
	}

	streamsDescriptions, err = c.getShardIterators(streamDescriptions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(streamsDescriptions))

}

func TestGetLatestRecords(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	shardIterator := aws.String("1")

	c := Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			getRecordsOutput:      dynamodbstreams.GetRecordsOutput{},
			getRecordsOutputError: errors.New("get records failed"),
		},
	}

	err := c.processLatestRecords(shardIterator)
	assert.Error(t, err)

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			getRecordsOutput: dynamodbstreams.GetRecordsOutput{
				Records: []*dynamodbstreams.Record{{EventID: aws.String("1")}},
			},
			getRecordsOutputError: nil,
		},
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:dynamodb",
				EventType: "DynamoDB update",
			},
		),
	}

	err = c.processLatestRecords(shardIterator)
	assert.NoError(t, err)

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			getRecordsOutput: dynamodbstreams.GetRecordsOutput{
				Records: []*dynamodbstreams.Record{{EventID: aws.String("1")}},
			},
			getRecordsOutputError: nil,
		},
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:dynamodb",
				EventType: "DynamoDB update",
			},
		),
	}

	err = c.processLatestRecords(shardIterator)
	assert.NoError(t, err)

	c = Clients{
		DynamoDBStream: mockedDynamoStreamsClient{
			getRecordsOutput: dynamodbstreams.GetRecordsOutput{
				Records: []*dynamodbstreams.Record{{EventID: aws.String("1")}},
			},
			getRecordsOutputError: nil,
		},
		CloudEvents: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:dynamodb",
				EventType: "DynamoDB update",
			},
		),
	}

	err = c.processLatestRecords(shardIterator)
	assert.NoError(t, err)

}

func TestSendCloudevent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	record := dynamodbstreams.Record{EventID: aws.String("1")}

	c := Clients{
		CloudEvents: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:dynamodb",
				EventType: "DynamoDB update",
			},
		),
	}

	err := c.sendDynamoDBEvent(&record)
	assert.Error(t, err)

	c = Clients{
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:dynamodb",
				EventType: "DynamoDB update",
			},
		),
	}

	err = c.sendDynamoDBEvent(&record)
	assert.NoError(t, err)
}
