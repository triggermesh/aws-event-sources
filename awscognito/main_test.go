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
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentity/cognitoidentityiface"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	"github.com/aws/aws-sdk-go/service/cognitosync/cognitosynciface"
	"github.com/jarcoal/httpmock"
	"github.com/knative/pkg/cloudevents"
	"github.com/stretchr/testify/assert"
)

type mockedCognitoIdentityClient struct {
	cognitoidentityiface.CognitoIdentityAPI
	listIdentitiesOutput      cognitoidentity.ListIdentitiesOutput
	listIdentitiesOutputError error
}

func (m mockedCognitoIdentityClient) ListIdentities(in *cognitoidentity.ListIdentitiesInput) (*cognitoidentity.ListIdentitiesOutput, error) {
	return &m.listIdentitiesOutput, m.listIdentitiesOutputError
}

type mockedCognitoSyncClient struct {
	cognitosynciface.CognitoSyncAPI
	listDatasetsOutput      cognitosync.ListDatasetsOutput
	listRecordsOutput       cognitosync.ListRecordsOutput
	listDatasetsOutputError error
	listRecordsOutputError  error
}

func (m mockedCognitoSyncClient) ListDatasets(in *cognitosync.ListDatasetsInput) (*cognitosync.ListDatasetsOutput, error) {
	return &m.listDatasetsOutput, m.listDatasetsOutputError
}

func (m mockedCognitoSyncClient) ListRecords(in *cognitosync.ListRecordsInput) (*cognitosync.ListRecordsOutput, error) {
	return &m.listRecordsOutput, m.listRecordsOutputError
}

func TestGetIdentities(t *testing.T) {
	clients := Clients{
		CognitoIdentity: mockedCognitoIdentityClient{
			listIdentitiesOutput:      cognitoidentity.ListIdentitiesOutput{},
			listIdentitiesOutputError: errors.New("ListIdentities failed"),
		},
	}
	_, err := clients.getIdentities()
	assert.Error(t, err)

	clients = Clients{
		CognitoIdentity: mockedCognitoIdentityClient{
			listIdentitiesOutput: cognitoidentity.ListIdentitiesOutput{
				Identities: []*cognitoidentity.IdentityDescription{{}, {}},
			},
			listIdentitiesOutputError: nil,
		},
	}
	identities, err := clients.getIdentities()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(identities))
}

func TestGetDatasets(t *testing.T) {
	identities := []*cognitoidentity.IdentityDescription{{IdentityId: aws.String("1")}}

	clients := Clients{
		CognitoSync: mockedCognitoSyncClient{
			listDatasetsOutput:      cognitosync.ListDatasetsOutput{},
			listDatasetsOutputError: errors.New("ListDatasets failed"),
		},
	}

	_, err := clients.getDatasets(identities)
	assert.Error(t, err)

	clients = Clients{
		CognitoSync: mockedCognitoSyncClient{
			listDatasetsOutput: cognitosync.ListDatasetsOutput{
				Datasets: []*cognitosync.Dataset{{}, {}},
			},
			listDatasetsOutputError: nil,
		},
	}

	datasets, err := clients.getDatasets(identities)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))

}
func TestGetRecords(t *testing.T) {
	dataset := cognitosync.Dataset{}

	clients := Clients{
		CognitoSync: mockedCognitoSyncClient{
			listRecordsOutput:      cognitosync.ListRecordsOutput{},
			listRecordsOutputError: errors.New("ListRecords failed"),
		},
	}

	_, err := clients.getRecords(&dataset)
	assert.Error(t, err)

	clients = Clients{
		CognitoSync: mockedCognitoSyncClient{
			listRecordsOutput: cognitosync.ListRecordsOutput{
				Records: []*cognitosync.Record{{}, {}},
			},
			listRecordsOutputError: nil,
		},
	}

	records, err := clients.getRecords(&dataset)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(records))
}
func TestSendCognitoEvent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://foo.com", httpmock.NewStringResponder(200, ``))
	httpmock.RegisterResponder("POST", "https://bar.com", httpmock.NewStringResponder(500, ``))

	dataset := cognitosync.Dataset{
		DatasetName: aws.String("foo"),
	}
	records := []*cognitosync.Record{}

	clients := Clients{
		CloudEvents: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:cognito",
				EventType: "SyncTrigger",
			},
		),
	}

	err := clients.sendCognitoEvent(&dataset, records)
	assert.Error(t, err)

	clients = Clients{
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:cognito",
				EventType: "SyncTrigger",
			},
		),
	}

	err = clients.sendCognitoEvent(&dataset, records)
	assert.NoError(t, err)
}
