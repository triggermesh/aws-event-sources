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
	"github.com/stretchr/testify/assert"
)

type mockedGetRecords struct {
	kinesisiface.KinesisAPI
	Resp kinesis.GetRecordsOutput
	err  error
}

func (m mockedGetRecords) GetRecords(in *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
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

	s := Stream{
		Client: mockedGetRecords{Resp: kinesis.GetRecordsOutput{
			NextShardIterator: aws.String("nextIterator"),
			Records:           records,
		}, err: nil},
	}

	inputs := []kinesis.GetRecordsInput{
		{},
	}

	err := s.processInputs(inputs)
	assert.NoError(t, err)

	s = Stream{
		Client: mockedGetRecords{Resp: kinesis.GetRecordsOutput{}, err: errors.New("error")},
	}

	inputs = []kinesis.GetRecordsInput{
		{},
	}

	err = s.processInputs(inputs)
	assert.NoError(t, err)

}

func TestSendCloudevent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))

	now := time.Now()
	record := kinesis.Record{
		Data: []byte("foo"),
		ApproximateArrivalTimestamp: &now,
		SequenceNumber:              aws.String("1"),
	}
	err := sendCloudevent(&record, "")
	assert.Error(t, err)

	err = sendCloudevent(&record, "https://foo.com")
	assert.NoError(t, err)
}
