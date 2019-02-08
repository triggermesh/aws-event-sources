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
	"flag"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	log "github.com/sirupsen/logrus"
	"github.com/triggermesh/sources/tmevents"
)

var (
	accountAccessKeyID     string
	accountSecretAccessKey string
	streamName             string
	region                 string
	sink                   string
)

// Stream provides the ability to operate on Kinesis stream.
type Stream struct {
	Client kinesisiface.KinesisAPI
	Stream *string
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {
	flag.Parse()

	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = os.Getenv("AWS_REGION")
	streamName = os.Getenv("STREAM")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}

	stream := Stream{kinesis.New(sess), &streamName}

	//Obtain records inputs for different shards
	inputs, err := stream.getRecordsInputs()
	if err != nil {
		log.Fatal(err)
	}

	for {
		err := stream.processInputs(inputs)
		if err != nil {
			log.Error(err)
		}
	}
}

func (s Stream) processInputs(inputs []kinesis.GetRecordsInput) error {

	for i, input := range inputs {

		recordsOutput, err := s.Client.GetRecords(&input)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, record := range recordsOutput.Records {
			err := sendCloudevent(record, sink)
			if err != nil {
				log.Errorf("SendCloudEvent failed: %v", err)
			}
		}

		//remove old imput
		inputs = append(inputs[:i], inputs[i+1:]...)

		//generate new input
		input = kinesis.GetRecordsInput{
			ShardIterator: recordsOutput.NextShardIterator,
		}

		//add newly generated input to the slice
		//so that new iteration would begin with new sharIterator
		inputs = append(inputs, input)
	}

	return nil
}

func (s Stream) getRecordsInputs() ([]kinesis.GetRecordsInput, error) {
	inputs := []kinesis.GetRecordsInput{}

	// Get info about a particular stream
	myStream, err := s.Client.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: s.Stream,
	})
	if err != nil {
		return inputs, err
	}

	//Kinesis stream might have several shards and each of them had "LATEST" Iterator.
	for _, shard := range myStream.StreamDescription.Shards {

		// Obtain starting Shard Iterator. This is needed to not process already processed records
		myShardIterator, err := s.Client.GetShardIterator(&kinesis.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("LATEST"),
			StreamName:        s.Stream,
		})

		if err != nil {
			log.Error(err)
			continue
		}

		// set records output limit. Should not be more than 10000, othervise panics
		input := kinesis.GetRecordsInput{
			ShardIterator: myShardIterator.ShardIterator,
		}

		inputs = append(inputs, input)
	}

	return inputs, err
}

func sendCloudevent(record *kinesis.Record, sink string) error {
	log.Info("Processing record ID: ", *record.SequenceNumber)

	eventInfo := tmevents.EventInfo{
		EventData:   record.Data,
		EventID:     *record.SequenceNumber,
		EventTime:   *record.ApproximateArrivalTimestamp,
		EventType:   "cloudevent.greet.you",
		EventSource: "aws-kinesis stream",
	}

	err := tmevents.PushEvent(&eventInfo, sink)
	if err != nil {
		return err
	}
	return nil
}
