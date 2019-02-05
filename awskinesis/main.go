package main

import (
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	log "github.com/sirupsen/logrus"
	"github.com/triggermesh/sources/tmevents"
)

var (
	accountAccessKeyID     string
	accountSecretAccessKey string
	stream                 string
	region                 string
	sink                   string

	shardIteratorType  string
	recordsOutputLimit int64
)

//KinesisStreamConsumer contains Kiness client and stream name
type KinesisStreamConsumer struct {
	client *kinesis.Kinesis
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")

	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	stream = os.Getenv("STREAM")
	region = os.Getenv("AWS_REGION")
	shardIteratorType = "LATEST"
	recordsOutputLimit = int64(10)

}

func main() {
	flag.Parse()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}

	consumer := KinesisStreamConsumer{kinesis.New(sess)}

	inputs, err := consumer.getInputs()
	if err != nil {
		log.Fatal(err)
	}

	for {
		for i, input := range inputs {

			recordsOutput, _, err := consumer.getRecords(input)
			if err != nil {
				log.Error(err)
				continue
			}

			//remove old imput
			inputs = append(inputs[:i], inputs[i+1:]...)

			//generate new input
			input = kinesis.GetRecordsInput{
				ShardIterator: recordsOutput.NextShardIterator,
				Limit:         &recordsOutputLimit,
			}

			//add newly generated input to the slice
			//so that new iteration would begin with new sharIterator
			inputs = append(inputs, input)

			if len(recordsOutput.Records) == 0 {
				log.Info("No records. Sleep for 10 seconds")
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func (c KinesisStreamConsumer) getInputs() ([]kinesis.GetRecordsInput, error) {
	inputs := []kinesis.GetRecordsInput{}

	// Get info about a particular stream
	myStream, err := c.client.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: &stream,
	})
	if err != nil {
		log.Error(err)
		return inputs, err
	}

	for _, shard := range myStream.StreamDescription.Shards {

		// Obtain starting Shard Iterator. This is needed to not process already processed records
		myShardIterator, err := c.client.GetShardIterator(&kinesis.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: &shardIteratorType,
			StreamName:        &stream,
		})

		if err != nil {
			log.Error(err)
			continue
		}

		// set records output limit. Should not be more than 10000, othervise panics
		input := kinesis.GetRecordsInput{
			ShardIterator: myShardIterator.ShardIterator,
			Limit:         &recordsOutputLimit,
		}

		inputs = append(inputs, input)
	}

	return inputs, err
}

func (c KinesisStreamConsumer) getRecords(input kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, []*kinesis.Record, error) {
	records := []*kinesis.Record{}
	recordsOutput, err := c.client.GetRecords(&input)
	if err != nil {
		log.Errorf("GetRecords failed: %v", err)
		return recordsOutput, records, err
	}

	for _, record := range recordsOutput.Records {
		records = append(records, record)
		go func(record *kinesis.Record) {
			err := sendCloudevent(record, sink)
			if err != nil {
				log.Errorf("SendCloudEvent failed: %v", err)
			}
		}(record)
	}

	return recordsOutput, records, nil
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
