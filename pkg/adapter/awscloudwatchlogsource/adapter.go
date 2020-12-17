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

package awscloudwatchlogsource

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN             string `envconfig:"ARN"`
	AccessKey       string `envconfig:"AWS_ACCESS_KEY_ID"`
	SecretKey       string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	PollingInterval string `envconfig:"POLLING_INTERVAL" required:"true"` // free tier is 5m
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	ceClient     cloudevents.Client
	cwLogsClient cloudwatchlogsiface.CloudWatchLogsAPI

	arn arn.ARN

	pollingInterval time.Duration
	logGroup        string
	logStream       string
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	var err error
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)
	a := common.MustParseARN(env.ARN)

	awsCfg := aws.NewConfig()
	awsCfg.WithCredentials(credentials.NewStaticCredentials(env.AccessKey, env.SecretKey, ""))
	cfg := session.Must(session.NewSession(awsCfg.
		WithRegion(a.Region).
		WithMaxRetries(5),
	))

	interval, err := time.ParseDuration(env.PollingInterval)
	if err != nil {
		logger.Panicf("Unable to parse interval duration: %v", zap.Error(err))
	}

	logGroup, logStream := ExtractLogDetails(a.Resource)

	return &adapter{
		logger: logger,

		pollingInterval: interval,
		logGroup:        logGroup,
		logStream:       logStream,

		cwLogsClient: cloudwatchlogs.New(cfg),
		ceClient:     ceClient,
		arn:          a,
	}
}

// ExtractLogDetails: Take the resource string from the ARN, and extract the `log-group` and `log-stream`
func ExtractLogDetails(details string) (string, string) {
	atoms := strings.Split(details, ":")

	var logGroup string
	var logStream string

	for i, k := range atoms {
		switch k {
		case "log-group":
			logGroup = atoms[i+1]
		case "log-stream":
			logStream = atoms[i+1]
		}
	}

	return logGroup, logStream
}

// Start implements adapter.Adapter.
func (a *adapter) Start(ctx context.Context) error {
	a.logger.Info("Enabling CloudWatchLog")

	// Setup polling to retrieve metrics
	poll := time.NewTicker(a.pollingInterval)
	metricsCh := make(chan bool)
	defer poll.Stop()
	defer close(metricsCh)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Cleanup thread
	go func() {
		<-ctx.Done()
		a.logger.Info("Shutdown signal received. Terminating")
		metricsCh <- true
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	// Primary execution thread. Wake up every pollintInterval, and retrieve the logs
	go func() {
		for {
			select {
			case <-metricsCh:
				wg.Done()
				return
			case t := <-poll.C:
				go a.CollectLogs(t)
			}
		}
	}()

	wg.Wait()
	return nil
}

func (a *adapter) CollectLogs(currentTime time.Time) {
	id := uuid.New() // In case multiple pages of log events need to be sent, keep the id consistent
	a.logger.Debug("Firing logs")
	startTime := currentTime.Add(-a.pollingInterval).Unix() * 1000
	endTime := currentTime.Unix() * 1000

	logStreams := cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: &a.logGroup,
	}

	err := a.cwLogsClient.DescribeLogStreamsPages(&logStreams, func(output *cloudwatchlogs.DescribeLogStreamsOutput, b bool) bool {
		var logRequest *cloudwatchlogs.GetLogEventsInput

		for _, v := range output.LogStreams {
			if v.LastIngestionTime != nil && *v.LastIngestionTime > startTime {

				if a.logStream != "" && a.logStream != "*" && *v.LogStreamName != a.logStream {
					continue
				}
				logRequest = &cloudwatchlogs.GetLogEventsInput{
					EndTime:       &endTime,
					LogGroupName:  &a.logGroup,
					LogStreamName: v.LogStreamName,
					StartTime:     &startTime,
				}
			} else {
				continue
			}

			page := 1 // Indicate number of pages of events
			err := a.cwLogsClient.GetLogEventsPages(logRequest, func(logOutput *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSCloudWatchLogsGenericEventType))
				event.SetSource(a.arn.String())
				event.SetID(id.String())

				// If there are no entries, then skip sending events
				if len(logOutput.Events) == 0 {
					a.logger.Debug("no log events sent")
					return !lastPage
				}

				// Ensure the entries captured within our range are the only events being published
				trimmedLogOutput := make([]*cloudwatchlogs.OutputLogEvent, 0)
				for _, v := range logOutput.Events {
					if *v.Timestamp >= startTime && *v.Timestamp < endTime {
						trimmedLogOutput = append(trimmedLogOutput, v)
					}
				}

				err := event.SetData(cloudevents.ApplicationJSON, logOutput.Events)
				if err != nil {
					a.logger.Errorf("failed to set event data: %v", err)
					return false
				}

				if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
					a.logger.Errorf("failed to send event data: %v", err)
					return false
				}

				page++
				return !lastPage
			})

			if err != nil {
				a.logger.Errorf("error retrieving logs: %v", zap.Error(err))
			}
		}

		return !b
	})

	if err != nil {
		a.logger.Errorf("error retrieving log streams: %v", zap.Error(err))
	}
}
