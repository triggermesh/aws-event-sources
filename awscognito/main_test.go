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
	c := Client{
		CognitoIdentity: mockedCognitoIdentityClient{
			listIdentitiesOutput:      cognitoidentity.ListIdentitiesOutput{},
			listIdentitiesOutputError: errors.New("ListIdentities failed"),
		},
	}
	_, err := c.getIdentities()
	assert.Error(t, err)

	c = Client{
		CognitoIdentity: mockedCognitoIdentityClient{
			listIdentitiesOutput: cognitoidentity.ListIdentitiesOutput{
				NextToken:  aws.String("next token"),
				Identities: []*cognitoidentity.IdentityDescription{{}, {}},
			},
			listIdentitiesOutputError: nil,
		},
	}
	identities, err := c.getIdentities()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(identities))

	c = Client{
		CognitoIdentity: mockedCognitoIdentityClient{
			listIdentitiesOutput: cognitoidentity.ListIdentitiesOutput{
				Identities: []*cognitoidentity.IdentityDescription{{}, {}},
			},
			listIdentitiesOutputError: nil,
		},
	}
	identities, err = c.getIdentities()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(identities))
}

func TestGetDatasets(t *testing.T) {
	identities := []*cognitoidentity.IdentityDescription{{IdentityId: aws.String("1")}}

	c := Client{
		CognitoSync: mockedCognitoSyncClient{
			listDatasetsOutput:      cognitosync.ListDatasetsOutput{},
			listDatasetsOutputError: errors.New("ListDatasets failed"),
		},
	}

	_, err := c.getDatasets(identities)
	assert.Error(t, err)

	c = Client{
		CognitoSync: mockedCognitoSyncClient{
			listDatasetsOutput: cognitosync.ListDatasetsOutput{
				Datasets: []*cognitosync.Dataset{{}, {}},
			},
			listDatasetsOutputError: nil,
		},
	}

	datasets, err := c.getDatasets(identities)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))

	c = Client{
		CognitoSync: mockedCognitoSyncClient{
			listDatasetsOutput: cognitosync.ListDatasetsOutput{
				NextToken: aws.String("next token"),
				Datasets:  []*cognitosync.Dataset{{}, {}},
			},
			listDatasetsOutputError: nil,
		},
	}

	datasets, err = c.getDatasets(identities)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))

}
func TestGetRecords(t *testing.T) {
	dataset := cognitosync.Dataset{}

	c := Client{
		CognitoSync: mockedCognitoSyncClient{
			listRecordsOutput:      cognitosync.ListRecordsOutput{},
			listRecordsOutputError: errors.New("ListRecords failed"),
		},
	}

	_, err := c.getRecords(&dataset)
	assert.Error(t, err)

	c = Client{
		CognitoSync: mockedCognitoSyncClient{
			listRecordsOutput: cognitosync.ListRecordsOutput{
				Records: []*cognitosync.Record{{}, {}},
			},
			listRecordsOutputError: nil,
		},
	}

	records, err := c.getRecords(&dataset)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(records))

	c = Client{
		CognitoSync: mockedCognitoSyncClient{
			listRecordsOutput: cognitosync.ListRecordsOutput{
				NextToken: aws.String("1"),
				Records:   []*cognitosync.Record{{}, {}},
			},
			listRecordsOutputError: nil,
		},
	}

	records, err = c.getRecords(&dataset)
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

	c := Client{
		CloudEvents: cloudevents.NewClient(
			"https://bar.com",
			cloudevents.Builder{
				Source:    "aws:cognito",
				EventType: "SyncTrigger",
			},
		),
	}

	err := c.sendCognitoEvent(&dataset, records)
	assert.Error(t, err)

	c = Client{
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:cognito",
				EventType: "SyncTrigger",
			},
		),
	}

	err = c.sendCognitoEvent(&dataset, records)
	assert.NoError(t, err)
}
