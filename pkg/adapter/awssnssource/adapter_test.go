/*
Copyright (c) 2019-2020 TriggerMesh Inc.

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

package awssnssource

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/stretchr/testify/assert"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
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
	subscribeURL, closeServer := mockServer()
	defer closeServer()

	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	data, err := ioutil.ReadFile("testSNSConfirmSubscriptionEvent.json")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	data = setSubscribeURL(t, data, subscribeURL)

	req, err := http.NewRequest("POST", "/", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.handleNotification)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	data, err = ioutil.ReadFile("testSNSNotificationEvent.json")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	req, err = http.NewRequest("POST", "/", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(a.handleNotification)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAttempSubscription(t *testing.T) {
	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	a.snsClient = mockedSNSClient{
		createTopicOutput:      sns.CreateTopicOutput{},
		createTopicOutputError: errors.New("fake error"),
	}

	err := a.attempSubscription(0)
	assert.Error(t, err)

	a.snsClient = mockedSNSClient{
		createTopicOutput:    sns.CreateTopicOutput{TopicArn: aws.String("fooArn")},
		subscribeOutput:      sns.SubscribeOutput{},
		subscribeOutputError: errors.New("fake error"),
	}

	err = a.attempSubscription(0)
	assert.Error(t, err)

	a.snsClient = mockedSNSClient{
		createTopicOutput:    sns.CreateTopicOutput{TopicArn: aws.String("fooArn")},
		subscribeOutput:      sns.SubscribeOutput{},
		subscribeOutputError: nil,
	}

	err = a.attempSubscription(0)
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

// mockServer starts a HTTP server and returns its URL together with a function
// to close it.
func mockServer() (string /*url*/, func() /*close*/) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return s.URL, s.Close
}

func setSubscribeURL(t *testing.T, data []byte, url string) []byte {
	structured := make(map[string]string)

	err := json.Unmarshal(data, &structured)
	if err != nil {
		t.Fatalf("Failed to deserialize data: %v", err)
	}

	structured["SubscribeURL"] = url

	data, err = json.Marshal(structured)
	if err != nil {
		t.Fatalf("Failed to serialize data: %v", err)
	}

	return data
}
