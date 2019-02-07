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
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	log "github.com/sirupsen/logrus"
	"github.com/triggermesh/sources/tmevents"
)

var (
	sink                   string
	accountAccessKeyID     string
	accountSecretAccessKey string
	region                 string
	identityPoolID         string
	maxResults             int64
)

func init() {
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = os.Getenv("AWS_REGION")
	identityPoolID = os.Getenv("IDENTITY_POOL_ID")
	maxResults = int64(10)

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

	ipConsumer := cognitoidentity.New(sess) // identity pool consumer
	syncConsumer := cognitosync.New(sess)

	descrIdentityPoolInput := cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: &identityPoolID,
	}
	identityPoolOutput, err := ipConsumer.DescribeIdentityPool(&descrIdentityPoolInput)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(identityPoolOutput.GoString())

	nextToken := "1"

	listIdentitiesInput := cognitoidentity.ListIdentitiesInput{
		IdentityPoolId: &identityPoolID,
		MaxResults:     &maxResults,
		NextToken:      &nextToken,
	}

	for {

		listIdentitiesOutput, err := ipConsumer.ListIdentities(&listIdentitiesInput)
		if err != nil {
			log.Error(err)
			continue
		}

		listIdentitiesInput.NextToken = listIdentitiesOutput.NextToken

		for _, identity := range listIdentitiesOutput.Identities {

			listDatasetsInput := cognitosync.ListDatasetsInput{
				IdentityPoolId: &identityPoolID,
				IdentityId:     identity.IdentityId,
				MaxResults:     &maxResults,
				NextToken:      &nextToken,
			}
			listDatasetsOutput, err := syncConsumer.ListDatasets(&listDatasetsInput)
			if err != nil {
				log.Fatal(err)
			}

			listDatasetsInput.NextToken = listDatasetsOutput.NextToken

			for _, dataset := range listDatasetsOutput.Datasets {
				go func(dataset *cognitosync.Dataset) {
					err := sendCloudevent(dataset, sink)
					if err != nil {
						log.Errorf("SendCloudEvent failed: %v", err)
					}
				}(dataset)
			}

		}
	}

}

func sendCloudevent(dataset *cognitosync.Dataset, sink string) error {
	log.Info("Processing Dataset: ", *dataset.DatasetName)

	eventInfo := tmevents.EventInfo{
		EventData:   []byte(dataset.String()),
		EventID:     *dataset.DatasetName,
		EventTime:   *dataset.CreationDate,
		EventType:   "cloudevent.greet.you",
		EventSource: "aws cognito",
	}

	err := tmevents.PushEvent(&eventInfo, sink)
	if err != nil {
		return err
	}
	return nil
}
