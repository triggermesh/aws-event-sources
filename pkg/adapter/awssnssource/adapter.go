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
	"sync"
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

	ARN                    string `required:"true"`
	SubscriptionAttributes []byte `split_words:"true"`
	PublicURL              string `required:"true" envconfig:"PUBLIC_URL"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	snsClient snsiface.SNSAPI
	ceClient  cloudevents.Client

	arn       arn.ARN
	subAttrs  map[string]*string
	publicURL url.URL
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	var subAttrs map[string]*string
	mustUnmarshalSubscriptionAttributes(env.SubscriptionAttributes, subAttrs)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region).
		WithMaxRetries(5),
	))

	return &adapter{
		logger: logger,

		snsClient: sns.New(cfg),
		ceClient:  ceClient,

		arn:       arn,
		subAttrs:  subAttrs,
		publicURL: mustParsePublicURL(env.PublicURL),
	}
}

func mustParsePublicURL(publicURL string) url.URL {
	url, err := url.Parse(publicURL)
	if err != nil {
		panic(err)
	}
	if url.String() == "" {
		panic("empty public URL")
	}

	return *url
}

func mustUnmarshalSubscriptionAttributes(data []byte, subAttrs map[string]*string) {
	if data == nil {
		return
	}

	err := json.Unmarshal(data, &subAttrs)
	if err != nil {
		panic(err)
	}
}

const (
	serverPort                = "8080"
	serverShutdownGracePeriod = time.Second * 10
	subscriptionRecheckPeriod = time.Second * 10
)

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	// ctx gets canceled to stop goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle stop signals
	go func() {
		<-stopCh
		a.logger.Info("Shutdown signal received. Terminating")
		cancel()
	}()

	http.HandleFunc("/", a.handleNotification)
	http.HandleFunc("/health", healthCheckHandler)

	server := &http.Server{Addr: ":" + serverPort}
	serverErrCh := make(chan error)
	defer close(serverErrCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		a.logger.Info("Serving on port " + serverPort)
		serverErrCh <- server.ListenAndServe()
		wg.Done()
	}()

	/* TODO(antoineco): we should delete the subscription when the source
	   is deleted by can't do it from the adapter because a) it should
	   scale to zero b) it shouldn't have access to the Kubernetes API to
	   read the event source object.
	   Ref. https://github.com/triggermesh/aws-event-sources/issues/157
	*/
	wg.Add(1)
	go func() {
		a.runSubscriptionReconciler(ctx, subscriptionRecheckPeriod)
		wg.Done()
	}()

	var err error

	select {
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			err = fmt.Errorf("failure during runtime of SNS notification handler: %w", serverErr)
		}
		cancel()

	case <-ctx.Done():
		a.logger.Info("Shutting server down")

		shutdownCtx, cancelTimeout := context.WithTimeout(ctx, serverShutdownGracePeriod)
		defer cancelTimeout()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			err = fmt.Errorf("error during server shutdown: %w", shutdownErr)
		}

		// unblock server goroutine
		<-serverErrCh
	}

	wg.Wait()
	return err
}

func (a *adapter) runSubscriptionReconciler(ctx context.Context, recheckPeriod time.Duration) {
	ticker := time.NewTicker(recheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.reconcileSubscription(); err != nil {
				a.logger.Errorw("Failed to reconcile Subscription", zap.Error(err))
			}
		case <-ctx.Done():
			a.logger.Info("Shutting subscription reconciler down")
			return
		}
	}
}

func (a *adapter) reconcileSubscription() error {
	resp, err := a.snsClient.Subscribe(&sns.SubscribeInput{
		Endpoint:   aws.String(a.publicURL.String()),
		Protocol:   &a.publicURL.Scheme,
		TopicArn:   aws.String(a.arn.String()),
		Attributes: a.subAttrs,
	})
	a.logger.Debug("Subscribe responded with: ", resp)

	return err
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

		a.logger.Debug("Successfully confirmed SNS subscription: ", *resp)

	// If the message is a notification, push the event
	// payload: https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html#http-notification-json
	case "Notification":
		event := cloudevents.NewEvent(cloudevents.VersionV1)
		event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSSNSGenericEventType))
		event.SetSource(a.arn.String())
		event.SetID(data["MessageId"].(string))

		if subjectAttr, ok := data["Subject"]; ok {
			event.SetSubject(subjectAttr.(string))
		}

		if err := event.SetData(cloudevents.ApplicationJSON, body); err != nil {
			a.logger.Errorw("Failed to set event data", zap.Error(err))
			http.Error(rw, fmt.Sprint("Failed to set event data: ", err), http.StatusInternalServerError)
			return
		}

		if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
			a.logger.Errorw("Failed to send CloudEvent", zap.Error(result))
			http.Error(rw, fmt.Sprint("Failed to send CloudEvent: ", result), http.StatusInternalServerError)
		}

		a.logger.Debug("Successfully sent SNS notification: ", event)
	}
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
