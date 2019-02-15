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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentity/cognitoidentityiface"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	"github.com/aws/aws-sdk-go/service/cognitosync/cognitosynciface"
	"github.com/knative/pkg/cloudevents"
	log "github.com/sirupsen/logrus"
)

var (
	sink                   string
	accountAccessKeyID     string
	accountSecretAccessKey string
	region                 string
	identityPoolID         string
)

//Client struct represent all clients
type Client struct {
	CognitoIdentity cognitoidentityiface.CognitoIdentityAPI
	CognitoSync     cognitosynciface.CognitoSyncAPI
	CloudEvents     *cloudevents.Client
}

//CognitoSyncEvent represents AWS CognitoSyncEvent payload
type CognitoSyncEvent struct {
	CreationDate     *time.Time
	DataStorage      *int64
	DatasetName      *string
	IdentityID       *string
	LastModifiedBy   *string
	LastModifiedDate *time.Time
	NumRecords       *int64
	EventType        *string
	Region           *string
	IdentityPoolID   *string
	DatasetRecords   []*cognitosync.Record
}

func init() {
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = os.Getenv("AWS_REGION")
	identityPoolID = os.Getenv("IDENTITY_POOL_ID")

	flag.StringVar(&sink, "sink", "", "where to sink events to")

}

func main() {

	flag.Parse()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	itentityClient := cognitoidentity.New(sess)
	syncClient := cognitosync.New(sess)
	cloudEvents := cloudevents.NewClient(
		sink,
		cloudevents.Builder{
			Source:    "aws:cognito",
			EventType: "SyncTrigger",
		},
	)

	client := Client{
		CognitoIdentity: itentityClient,
		CognitoSync:     syncClient,
		CloudEvents:     cloudEvents,
	}

	for {

		identities, err := client.getIdentities()
		if err != nil {
			log.Error(err)
		}

		datasets, err := client.getDatasets(identities)
		if err != nil {
			log.Error(err)
		}

		for _, dataset := range datasets {
			records, err := client.getRecords(dataset)
			if err != nil {
				log.Error(err)
				continue
			}

			err = client.sendCognitoEvent(dataset, records)
			if err != nil {
				log.Errorf("SendCloudEvent failed: %v", err)
			}
		}
	}
}

func (client Client) getIdentities() ([]*cognitoidentity.IdentityDescription, error) {
	identities := []*cognitoidentity.IdentityDescription{}

	listIdentitiesInput := cognitoidentity.ListIdentitiesInput{
		IdentityPoolId: &identityPoolID,
	}

	for {
		listIdentitiesOutput, err := client.CognitoIdentity.ListIdentities(&listIdentitiesInput)
		if err != nil {
			return identities, err
		}

		identities = append(identities, listIdentitiesOutput.Identities...)

		if listIdentitiesOutput.NextToken == nil {
			break
		}

		listIdentitiesInput.NextToken = listIdentitiesOutput.NextToken
	}

	return identities, nil
}

func (client Client) getDatasets(identities []*cognitoidentity.IdentityDescription) ([]*cognitosync.Dataset, error) {
	datasets := []*cognitosync.Dataset{}

	for _, identity := range identities {
		listDatasetsInput := cognitosync.ListDatasetsInput{
			IdentityPoolId: &identityPoolID,
			IdentityId:     identity.IdentityId,
		}

		for {
			listDatasetsOutput, err := client.CognitoSync.ListDatasets(&listDatasetsInput)
			if err != nil {
				return datasets, err
			}

			datasets = append(datasets, listDatasetsOutput.Datasets...)

			if listDatasetsOutput.NextToken == nil {
				break
			}

			listDatasetsInput.NextToken = listDatasetsOutput.NextToken
		}
	}

	return datasets, nil
}

func (client Client) getRecords(dataset *cognitosync.Dataset) ([]*cognitosync.Record, error) {
	records := []*cognitosync.Record{}

	input := cognitosync.ListRecordsInput{
		DatasetName:    dataset.DatasetName,
		IdentityId:     dataset.IdentityId,
		IdentityPoolId: aws.String(identityPoolID),
	}

	for {
		recordsOutput, err := client.CognitoSync.ListRecords(&input)
		if err != nil {
			return records, err
		}

		records = append(records, recordsOutput.Records...)

		if recordsOutput.NextToken == nil {
			break
		}

		input.NextToken = recordsOutput.NextToken
	}

	return records, nil
}

func (client Client) sendCognitoEvent(dataset *cognitosync.Dataset, records []*cognitosync.Record) error {
	log.Info("Processing Dataset: ", *dataset.DatasetName)

	cognitoEvent := CognitoSyncEvent{
		CreationDate:     dataset.CreationDate,
		DataStorage:      dataset.DataStorage,
		DatasetName:      dataset.DatasetName,
		IdentityID:       dataset.IdentityId,
		LastModifiedBy:   dataset.LastModifiedBy,
		LastModifiedDate: dataset.LastModifiedDate,
		NumRecords:       dataset.NumRecords,
		EventType:        aws.String("SyncTrigger"),
		Region:           aws.String(region),
		IdentityPoolID:   aws.String(identityPoolID),
		DatasetRecords:   records,
	}

	if err := client.CloudEvents.Send(cognitoEvent); err != nil {
		return err
	}

	return nil
}
