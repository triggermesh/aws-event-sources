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

package awscognitosource

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentity/cognitoidentityiface"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	"github.com/aws/aws-sdk-go/service/cognitosync/cognitosynciface"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	IdentityPoolId         string `envconfig:"IDENTITY_POOL_ID" required:"true"`
	AccountAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AccountSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	cgnIdentityClient cognitoidentityiface.CognitoIdentityAPI
	cgnSyncClient     cognitosynciface.CognitoSyncAPI
	ceClient          cloudevents.Client

	identityPoolId         string
	awsRegion              string
	accountAccessKeyId     string
	accountSecretAccessKey string
}

func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor,
	ceClient cloudevents.Client) pkgadapter.Adapter {

	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	region := extractRegionFromPoolId(env.IdentityPoolId)

	// create Cognito clients
	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(env.AccountAccessKeyId, env.AccountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logger.Fatalw("Failed to create Cognito clients", "error", err)
	}

	return &adapter{
		logger: logger,

		cgnIdentityClient: cognitoidentity.New(sess),
		cgnSyncClient:     cognitosync.New(sess),
		ceClient:          ceClient,

		identityPoolId:         env.IdentityPoolId,
		awsRegion:              region,
		accountAccessKeyId:     env.AccountAccessKeyId,
		accountSecretAccessKey: env.AccountSecretAccessKey,
	}
}

// extractRegionFromPoolId parses an identity pool id and returns the AWS region.
// TODO(antoineco): consolidate this duplicated function (already used in the adapter)
func extractRegionFromPoolId(identityPoolId string) (region string) {
	subs := strings.Split(identityPoolId, ":")
	// extra safety, the API validation should have already ensured that the
	// format of the id is correct
	if len(subs) == 2 {
		region = subs[0]
	}
	return
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	a.logger.Infof("Listening to AWS Cognito stream for Identity: %s", a.identityPoolId)

	for {
		identities, err := a.getIdentities()
		if err != nil {
			a.logger.Error(err)
		}

		datasets, err := a.getDatasets(identities)
		if err != nil {
			a.logger.Error(err)
		}

		for _, dataset := range datasets {
			records, err := a.getRecords(dataset)
			if err != nil {
				a.logger.Error(err)
				continue
			}

			err = a.sendCognitoEvent(dataset, records)
			if err != nil {
				a.logger.Errorf("SendCloudEvent failed: %v", err)
			}
		}
	}
}

func (a *adapter) getIdentities() ([]*cognitoidentity.IdentityDescription, error) {
	identities := []*cognitoidentity.IdentityDescription{}

	listIdentitiesInput := cognitoidentity.ListIdentitiesInput{
		MaxResults:     aws.Int64(1),
		IdentityPoolId: &a.identityPoolId,
	}

	for {
		listIdentitiesOutput, err := a.cgnIdentityClient.ListIdentities(&listIdentitiesInput)
		if err != nil {
			return identities, err
		}

		identities = append(identities, listIdentitiesOutput.Identities...)

		listIdentitiesInput.NextToken = listIdentitiesOutput.NextToken
		if listIdentitiesOutput.NextToken == nil {
			break
		}

	}

	return identities, nil
}

func (a *adapter) getDatasets(identities []*cognitoidentity.IdentityDescription) ([]*cognitosync.Dataset, error) {
	datasets := []*cognitosync.Dataset{}

	for _, identity := range identities {
		listDatasetsInput := cognitosync.ListDatasetsInput{
			IdentityPoolId: &a.identityPoolId,
			IdentityId:     identity.IdentityId,
		}

		for {
			listDatasetsOutput, err := a.cgnSyncClient.ListDatasets(&listDatasetsInput)
			if err != nil {
				return datasets, err
			}

			datasets = append(datasets, listDatasetsOutput.Datasets...)

			listDatasetsInput.NextToken = listDatasetsOutput.NextToken
			if listDatasetsOutput.NextToken == nil {
				break
			}
		}
	}

	return datasets, nil
}

func (a *adapter) getRecords(dataset *cognitosync.Dataset) ([]*cognitosync.Record, error) {
	records := []*cognitosync.Record{}

	input := cognitosync.ListRecordsInput{
		DatasetName:    dataset.DatasetName,
		IdentityId:     dataset.IdentityId,
		IdentityPoolId: &a.identityPoolId,
	}

	for {
		recordsOutput, err := a.cgnSyncClient.ListRecords(&input)
		if err != nil {
			return records, err
		}

		records = append(records, recordsOutput.Records...)

		input.NextToken = recordsOutput.NextToken
		if recordsOutput.NextToken == nil {
			break
		}
	}

	return records, nil
}

func (a *adapter) sendCognitoEvent(dataset *cognitosync.Dataset, records []*cognitosync.Record) error {
	a.logger.Info("Processing Dataset: ", *dataset.DatasetName)

	data := &CognitoSyncEvent{
		CreationDate:     dataset.CreationDate,
		DataStorage:      dataset.DataStorage,
		DatasetName:      dataset.DatasetName,
		IdentityID:       dataset.IdentityId,
		LastModifiedBy:   dataset.LastModifiedBy,
		LastModifiedDate: dataset.LastModifiedDate,
		NumRecords:       dataset.NumRecords,
		EventType:        aws.String("SyncTrigger"),
		Region:           &a.awsRegion,
		IdentityPoolID:   &a.identityPoolId,
		DatasetRecords:   records,
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(v1alpha1.AWSCognitoEventType(v1alpha1.AWSCognitoGenericEventType))
	event.SetSubject(a.identityPoolId)
	event.SetSource(v1alpha1.AWSCognitoEventSource(a.identityPoolId))
	event.SetID(*dataset.IdentityId)
	event.SetData(cloudevents.ApplicationJSON, data)

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
