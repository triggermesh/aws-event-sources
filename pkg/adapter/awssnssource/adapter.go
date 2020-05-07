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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN string `envconfig:"ARN" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	snsClient snsiface.SNSAPI
	ceClient  cloudevents.Client

	arn arn.ARN
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region).
		WithMaxRetries(5),
	))

	return &adapter{
		logger: logger,

		snsClient: sns.New(cfg),
		ceClient:  ceClient,

		arn: arn,
	}
}

const (
	port                      = 8081
	defaultSubscriptionPeriod = 10 * time.Second
)

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	// Setup subscription in the background. Will keep us from having chicken/egg between server
	// being ready to respond and us having the info we need for the subscription request
	go func() {
		for {
			if err := a.attempSubscription(defaultSubscriptionPeriod); err != nil {
				a.logger.Error(err)
			}
		}
	}()

	// Start server
	http.HandleFunc("/", a.handleNotification)
	http.HandleFunc("/health", healthCheckHandler)
	a.logger.Infof("Serving on port %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (a *adapter) attempSubscription(period time.Duration) error {
	time.Sleep(period)

	topic, err := a.snsClient.CreateTopic(&sns.CreateTopicInput{Name: &a.arn.Resource})
	if err != nil {
		return err
	}

	sink := os.Getenv("K_SINK")
	sinkURL, err := url.Parse(sink)
	if err != nil {
		return err
	}

	_, err = a.snsClient.Subscribe(&sns.SubscribeInput{
		Endpoint: &sink,
		Protocol: &sinkURL.Scheme,
		TopicArn: topic.TopicArn,
	})
	if err != nil {
		return err
	}

	a.logger.Debug("Finished subscription flow")
	return nil
}

// handleNotification implements the receive interface for SNS.
func (a *adapter) handleNotification(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.logger.Errorw("Failed to read request body", zap.Error(err))
		http.Error(rw, fmt.Sprint("Failed to read request body: ", err), http.StatusInternalServerError)
		return
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(body, &data); err != nil {
		a.logger.Errorw("Failed to parse notification", zap.Error(err))
		http.Error(rw, fmt.Sprint("Failed to parse notification: ", err), http.StatusBadRequest)
		return
	}

	a.logger.Debug("Request body: ", string(body))

	switch data["Type"].(string) {
	// If the message is about our subscription, call the confirmation endpoint.
	// payload: https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html#http-subscription-confirmation-json
	case "SubscriptionConfirmation":
		resp, err := a.snsClient.ConfirmSubscription(&sns.ConfirmSubscriptionInput{
			TopicArn: aws.String(data["TopicArn"].(string)),
			Token:    aws.String(data["Token"].(string)),
		})
		if err != nil {
			a.logger.Errorw("Unable to confirm SNS subscription", zap.Error(err))
			http.Error(rw, fmt.Sprint("Unable to confirm SNS subscription: ", err), http.StatusInternalServerError)
			return
		}

		a.logger.Debug("Successfully confirmed SNS subscription ", *resp.SubscriptionArn)

	// If the message is a notification, push the event
	// payload: https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html#http-notification-json
	case "Notification":
		eventTime, err := time.Parse(time.RFC3339, data["Timestamp"].(string))
		if err != nil {
			a.logger.Errorw("Failed to parse notification timestamp", zap.Error(err))
			http.Error(rw, fmt.Sprint("Failed to parse notification timestamp: ", err), http.StatusBadRequest)
			return
		}

		record := &SNSEventRecord{
			EventVersion: "1.0",
			EventSource:  "aws:sns",
			SNS: SNSEntity{
				Message:          data["Message"].(string),
				MessageID:        data["MessageId"].(string),
				Signature:        data["Signature"].(string),
				SignatureVersion: data["SignatureVersion"].(string),
				SigningCertURL:   data["SigningCertURL"].(string),
				Subject:          data["Subject"].(string),
				Timestamp:        eventTime,
				TopicArn:         data["TopicArn"].(string),
				Type:             data["Type"].(string),
				UnsubscribeURL:   data["UnsubscribeURL"].(string),
			},
		}

		event := cloudevents.NewEvent(cloudevents.VersionV1)
		event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSSNSGenericEventType))
		event.SetSubject(data["Subject"].(string))
		event.SetSource(a.arn.String())
		event.SetID(data["MessageId"].(string))
		if err := event.SetData(cloudevents.ApplicationJSON, record); err != nil {
			a.logger.Errorw("Failed to set event data", zap.Error(err))
			http.Error(rw, fmt.Sprint("Failed to set event data: ", err), http.StatusInternalServerError)
			return
		}

		if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
			a.logger.Errorw("Failed to send CloudEvent", "error", result)
			http.Error(rw, fmt.Sprint("Failed to send CloudEvent: ", result), http.StatusInternalServerError)
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
