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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	Topic                  string `envconfig:"TOPIC" required:"true"`
	AWSRegion              string `envconfig:"AWS_REGION" required:"true"`
	AccountAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AccountSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	snsClient snsiface.SNSAPI
	ceClient  cloudevents.Client

	topic                  string
	awsRegion              string
	accountAccessKeyID     string
	accountSecretAccessKey string
}

func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor,
	ceClient cloudevents.Client) pkgadapter.Adapter {

	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	// create SNS client
	sess, err := session.NewSession(&aws.Config{
		Region:      &env.AWSRegion,
		Credentials: credentials.NewStaticCredentials(env.AccountAccessKeyId, env.AccountSecretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		logger.Fatalw("Failed to create SNS client", "error", err)
	}

	return &adapter{
		logger: logger,

		snsClient: sns.New(sess),
		ceClient:  ceClient,

		topic:                  env.Topic,
		awsRegion:              env.AWSRegion,
		accountAccessKeyID:     env.AccountAccessKeyId,
		accountSecretAccessKey: env.AccountSecretAccessKey,
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

	topic, err := a.snsClient.CreateTopic(&sns.CreateTopicInput{Name: &a.topic})
	if err != nil {
		return err
	}

	sink := os.Getenv("K_SINK")
	sinkUrl, err := url.Parse(sink)
	if err != nil {
		return err
	}

	_, err = a.snsClient.Subscribe(&sns.SubscribeInput{
		Endpoint: &sink,
		Protocol: &sinkUrl.Scheme,
		TopicArn: topic.TopicArn,
	})
	if err != nil {
		return err
	}

	a.logger.Debug("Finished subscription flow")
	return nil
}

// handleNotification implements the receive interface for SNS.
func (a *adapter) handleNotification(_ http.ResponseWriter, r *http.Request) {
	// Fish out notification body
	var notification interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.logger.Error("Failed to parse notification: ", err)
	}
	err = json.Unmarshal(body, &notification)
	if err != nil {
		a.logger.Error("Failed to parse notification: ", err)
	}
	a.logger.Info(string(body))
	data := notification.(map[string]interface{})

	// If the message is about our subscription, curl the confirmation endpoint.
	if data["Type"].(string) == "SubscriptionConfirmation" {

		subcribeURL := data["SubscribeURL"].(string)
		_, err := http.Get(subcribeURL)
		if err != nil {
			a.logger.Fatalw("Unable to confirm SNS subscription", "error", err)
		}
		a.logger.Info("Successfully confirmed SNS subscription")

		// If it's a legit notification, push the event
	} else if data["Type"].(string) == "Notification" {

		eventTime, _ := time.Parse(time.RFC3339, data["Timestamp"].(string))

		record := &SNSEventRecord{
			EventVersion:         "1.0",
			EventSubscriptionArn: "",
			EventSource:          "aws:sns",
			SNS: SNSEntity{
				Signature:         data["Signature"].(string),
				MessageID:         data["MessageId"].(string),
				Type:              data["Type"].(string),
				TopicArn:          data["TopicArn"].(string),
				MessageAttributes: data["MessageAttributes"].(map[string]interface{}),
				SignatureVersion:  data["SignatureVersion"].(string),
				Timestamp:         eventTime,
				SigningCertURL:    data["SigningCertURL"].(string),
				Message:           data["Message"].(string),
				UnsubscribeURL:    data["UnsubscribeURL"].(string),
				Subject:           data["Subject"].(string),
			},
		}

		event := cloudevents.NewEvent(cloudevents.VersionV1)
		event.SetType(v1alpha1.AWSSNSEventType(v1alpha1.AWSKinesisGenericEventType))
		event.SetSubject(data["Subject"].(string))
		event.SetSource(v1alpha1.AWSSNSEventSource(a.awsRegion, a.topic))
		event.SetID(data["MessageId"].(string))
		event.SetData(cloudevents.ApplicationJSON, record)

		if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
			a.logger.Errorw("Failed to send CloudEvent", "error", err)
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("OK"))
}
