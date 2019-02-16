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
	"bufio"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/jarcoal/httpmock"
	"github.com/knative/pkg/cloudevents"
	"github.com/stretchr/testify/assert"
)

type mockedSNSClient struct {
	snsiface.SNSAPI
	createTopicOutput      sns.CreateTopicOutput
	createTopicOutputError error

	subscribeOutput      sns.SubscribeOutput
	subscribeOutputError error
}

func (m mockedSNSClient) CreateTopic(in *sns.CreateTopicInput) (*sns.CreateTopicOutput, error) {
	return &m.createTopicOutput, m.createTopicOutputError
}

func (m mockedSNSClient) Subscribe(in *sns.SubscribeInput) (*sns.SubscribeOutput, error) {
	return &m.subscribeOutput, m.subscribeOutputError
}

func TestHandleNotification(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://foo.com", httpmock.NewStringResponder(200, ``))

	client := Client{
		CloudEvents: cloudevents.NewClient(
			"https://foo.com",
			cloudevents.Builder{
				Source:    "aws:sns",
				EventType: "SNS Event",
			},
		),
	}

	file, err := os.Open("testSNSConfirmSubscriptionEvent.json")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bufio.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(client.HandleNotification)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	file, err = os.Open("testSNSNotificationEvent.json")
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/", bufio.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(client.HandleNotification)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAttempSubscription(t *testing.T) {

	c := Client{
		SNSClient: mockedSNSClient{
			createTopicOutput:      sns.CreateTopicOutput{},
			createTopicOutputError: errors.New("err"),
		},
	}

	err := c.attempSubscription()
	assert.Error(t, err)

	c = Client{
		SNSClient: mockedSNSClient{
			createTopicOutput:    sns.CreateTopicOutput{TopicArn: aws.String("fooArn")},
			subscribeOutput:      sns.SubscribeOutput{},
			subscribeOutputError: errors.New("err"),
		},
	}

	err = c.attempSubscription()
	assert.Error(t, err)

	c = Client{
		SNSClient: mockedSNSClient{
			createTopicOutput:    sns.CreateTopicOutput{TopicArn: aws.String("fooArn")},
			subscribeOutput:      sns.SubscribeOutput{},
			subscribeOutputError: nil,
		},
	}

	err = c.attempSubscription()
	assert.NoError(t, err)
}

func TestHealth(t *testing.T) {

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
