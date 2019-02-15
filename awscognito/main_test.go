package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentity/cognitoidentityiface"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	"github.com/aws/aws-sdk-go/service/cognitosync/cognitosynciface"
)

type mockedCognitoIdentityClient struct {
	cognitoidentityiface.CognitoIdentityAPI
	listIdentitiesOutput cognitoidentity.ListIdentitiesOutput
	err                  error
}

func (m mockedCognitoIdentityClient) ListIdentities(in *cognitoidentity.ListIdentitiesInput) (*cognitoidentity.ListIdentitiesOutput, error) {
	return &m.listIdentitiesOutput, m.err
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

}

func TestGetDatasets(t *testing.T) {

}
func TestGetRecords(t *testing.T) {

}
func TestSendCognitoEvent(t *testing.T) {

}
