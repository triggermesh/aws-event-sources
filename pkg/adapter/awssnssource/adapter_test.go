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
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"
)

type mockedSNSClient struct {
	snsiface.SNSAPI

	subscribeOutput      *sns.SubscribeOutput
	subscribeOutputError error

	confirmSubsOutput *sns.ConfirmSubscriptionOutput
	confirmSubsError  error
}

func (m mockedSNSClient) Subscribe(_ *sns.SubscribeInput) (*sns.SubscribeOutput, error) {
	return m.subscribeOutput, m.subscribeOutputError
}

func (m mockedSNSClient) ConfirmSubscription(_ *sns.ConfirmSubscriptionInput) (*sns.ConfirmSubscriptionOutput, error) {
	return m.confirmSubsOutput, m.confirmSubsError
}

// TestStart verifies that a started adapter responds to cancelation.
func TestStart(t *testing.T) {
	const testTimeout = time.Second * 2
	testCtx, testCancel := context.WithTimeout(context.Background(), testTimeout)
	defer testCancel()

	a := &adapter{
		logger: loggingtesting.TestLogger(t),
	}

	// errCh receives the error value returned by the receiver after
	// termination. We leave it open to avoid panicking in case the
	// receiver returns after the timeout.
	errCh := make(chan error)

	// ctx gets canceled to cause a voluntary interruption of the receiver
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		errCh <- a.Start(ctx)
	}()
	cancel()

	select {
	case <-testCtx.Done():
		t.Errorf("Test timed out after %v", testTimeout)
	case err := <-errCh:
		if err != nil {
			t.Errorf("Receiver returned an error: %s", err)
		}
	}
}

func TestHandler(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	a.snsClient = mockedSNSClient{
		confirmSubsOutput: &sns.ConfirmSubscriptionOutput{SubscriptionArn: aws.String("fooArn")},
	}

	// handle subscribe

	data, err := ioutil.ReadFile("testSNSConfirmSubscriptionEvent.json")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.handleNotification)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	gotEvents := ceClient.Sent()
	assert.Len(t, gotEvents, 0, "Expect no event")

	// handle notification

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

	gotEvents = ceClient.Sent()
	assert.Len(t, gotEvents, 1, "Expect a single event")

	gotData := gotEvents[0].Data()
	assert.EqualValues(t, data, gotData, "Received event data should equal sent payload")
}

func TestReconcileSubscription(t *testing.T) {
	a := &adapter{
		logger:    loggingtesting.TestLogger(t),
		publicURL: url.URL{Scheme: "http", Host: "example.com"},
	}

	a.snsClient = mockedSNSClient{
		subscribeOutput:      &sns.SubscribeOutput{},
		subscribeOutputError: errors.New("fake error"),
	}

	err := a.reconcileSubscription()
	assert.Error(t, err)

	a.snsClient = mockedSNSClient{
		subscribeOutput:      &sns.SubscribeOutput{},
		subscribeOutputError: nil,
	}

	err = a.reconcileSubscription()
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
	assert.Equal(t, "OK\n", rr.Body.String())
}
