package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/sirupsen/logrus"
)

var (
	accountAccessKeyID     string
	accountSecretAccessKey string

	stream    string
	region    string
	namespace string
	channel   string
)

func main() {

	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	stream = os.Getenv("STREAM")
	region = os.Getenv("AWS_REGION")

	mySession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logrus.Fatal(err)
	}

	// Create a Kinesis client with additional configuration
	svc := kinesis.New(mySession, aws.NewConfig().WithRegion(region))

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

		_, err := svc.PutRecord(&myRecord)

		if err != nil {
			logrus.Error("PutRecord failed: ", err)
		}
		logrus.Info("record inserted!")
	}

}
