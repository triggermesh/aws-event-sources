package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/sirupsen/logrus"
)

func TestGetShards(t *testing.T) {
	stream := "teststream"
	shardCount := int64(2)
	region = "us-west-2"
	enforceConsumerDeletion := true

	mySession, err := session.NewSession(aws.NewConfig())
	if err != nil {
		logrus.Fatal(err)
	}

	c := KinesisStreamConsumer{
		client: kinesis.New(mySession, aws.NewConfig().WithRegion(region)),
	}

	createStreamInput := kinesis.CreateStreamInput{
		ShardCount: &shardCount,
		StreamName: &stream,
	}
	logrus.Info("Creating Stream")
	_, err = c.client.CreateStream(&createStreamInput)
	if err != nil {
		logrus.Error(err)
	}

	time.Sleep(20 * time.Second) // need to wait before stream creates

	for i := 0; i <= 10; i++ {
		myRecord := kinesis.PutRecordInput{}
		myRecord.SetData([]byte(fmt.Sprintf("Record #%v", i)))
		//to get 50% of data into a different shard
		if i%2 == 0 {
			myRecord.SetExplicitHashKey("170141183460469231731687303715884105729")
			myRecord.SetPartitionKey("test2ndShard")
		} else {
			myRecord.SetPartitionKey("testKey")
		}

		myRecord.SetStreamName(stream)

		_, err := c.client.PutRecord(&myRecord)

		if err != nil {
			logrus.Error("PutRecord failed: ", err)
		}
		logrus.Info("record inserted!")
	}

	inputs, err := c.getInputs()
	if err != nil {
		t.Error(err)
	}

	if int64(len(inputs)) != shardCount {
		t.Errorf("Wrong number of inputs in the stream. Expecting %v, got %v", shardCount, int64(len(inputs)))
	}

	_, recordsFromFirstShard, err := c.getRecords(inputs[0])

	if err != nil {
		t.Error(err)
	}

	if len(recordsFromFirstShard) != 5 {
		t.Errorf("Wrong number of records in the shard. Expecting %v, got %v", 5, len(recordsFromFirstShard))
	}

	_, recordsFromSecondShard, err := c.getRecords(inputs[1])

	if err != nil {
		t.Error(err)
	}

	if len(recordsFromSecondShard) != 5 {
		t.Errorf("Wrong number of records in the shard. Expecting %v, got %v", 5, len(recordsFromSecondShard))
	}

	deleteStreamInput := kinesis.DeleteStreamInput{
		EnforceConsumerDeletion: &enforceConsumerDeletion,
		StreamName:              &stream,
	}
	logrus.Info("Deleting Stream")
	_, err = c.client.DeleteStream(&deleteStreamInput)
	if err != nil {
		t.Error(err)
	}

}
