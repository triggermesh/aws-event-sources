package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	log "github.com/sirupsen/logrus"
)

var sink string

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	accountAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	accountRegion := os.Getenv("AWS_REGION")
	myBucket := os.Getenv("AWS_BUCKET")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(accountRegion),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
	})
	if err != nil {
		log.Errorf("NewSession failed: %v", err)
		return
	}

	svc := s3.New(sess)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(myBucket),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range result.Contents {
		connectToObjectEvent(svc, myBucket, object.Key)
	}

}

func connectToObjectEvent(svc *s3.S3, bucket string, objectKey *string) {
	resp, err := svc.SelectObjectContent(&s3.SelectObjectContentInput{
		Bucket:         aws.String(bucket),
		Key:            objectKey,
		Expression:     aws.String("SELECT * FROM S3Object"),
		ExpressionType: aws.String(s3.ExpressionTypeSql),
		InputSerialization: &s3.InputSerialization{
			CSV: &s3.CSVInput{
				FileHeaderInfo: aws.String(s3.FileHeaderInfoUse),
			},
		},
		OutputSerialization: &s3.OutputSerialization{
			JSON: &s3.JSONOutput{},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed making API request, %v\n", err)
		return
	}

	log.Info(resp)
}
