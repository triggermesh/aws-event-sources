/*
Copyright (c) 2021 TriggerMesh Inc.

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

package awsperformanceinsightssource

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pi"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN string `envconfig:"ARN"`

	PollingInterval string `envconfig:"POLLING_INTERVAL" required:"true"`

	MetricQuerys []string `envconfig:"METRIC_QUERYS" required:"true"`

	Identifier string `envconfig:"IDENTIFIER" required:"true"`

	ServiceType string `envconfig:"SERVICE_TYPE" required:"true" `
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	pIClient *pi.PI
	ceClient cloudevents.Client

	arn             arn.ARN
	pollingInterval time.Duration
	metricQuerys    []*pi.MetricQuery
	identifier      string
	serviceType     string
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

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(a.Region),
	))

	interval, err := time.ParseDuration(env.PollingInterval)
	if err != nil {
		logger.Panicf("Unable to parse interval duration: %v", zap.Error(err))
	}

	// var testMQL []string

	// testMQL = append(testMQL, "os.cpuUtilization.idle.avg")
	// testMQL = append(testMQL, "os.general.numVCPUs.avg")
	// testMQL = append(testMQL, "os.network.rx.avg")
	// testMQL = append(testMQL, "os.network.tx.avg")

	// testMQL = append(testMQL, "os.network.rx.avg")

	var mql []*pi.MetricQuery

	for _, r := range env.MetricQuerys {
		m := &pi.MetricQuery{Metric: aws.String(r)}
		mql = append(mql, m)
	}

	return &adapter{
		logger: logger,

		pIClient: pi.New(cfg),
		ceClient: ceClient,

		arn: a,

		pollingInterval: interval,
		metricQuerys:    mql,
		identifier:      env.Identifier,
		serviceType:     env.ServiceType,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(ctx context.Context) error {
	a.logger.Info("Enabling AWS Performance Insights Source")

	// Setup polling to retrieve metrics
	poll := time.NewTicker(a.pollingInterval)
	defer poll.Stop()

	// Wake up every pollingInterval, and retrieve the logs
	var priorTime time.Time
	priorTime = time.Now()
	for {
		select {
		case <-ctx.Done():
			return nil

		case t := <-poll.C:
			go a.PollMetrics(priorTime, t)
			priorTime = t
		}
	}
}

func (a *adapter) PollMetrics(priorTime time.Time, currentTime time.Time) {
	rmi := &pi.GetResourceMetricsInput{
		EndTime:       aws.Time(time.Now()),
		StartTime:     aws.Time(priorTime),
		Identifier:    aws.String(a.identifier),
		MetricQueries: a.metricQuerys,
		ServiceType:   aws.String(a.serviceType),
	}

	rm, err := a.pIClient.GetResourceMetrics(rmi)

	if err != nil {
		a.logger.Errorf("retrieving resource metrics: %v", err)
		return
	}
	fmt.Println(rm)
	for _, d := range rm.MetricList {
		for _, metric := range d.DataPoints {
			if metric.Value != nil {
				e := &event{
					Metric: *d.Key.Metric,
					Value:  *metric.Value,
				}

				fmt.Println("________________")
				fmt.Println(e)
				fmt.Println("________________")

				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSPerformanceInsightsGenericEventType))
				event.SetSource(*d.Key.Metric)

				ceer := event.SetData(cloudevents.ApplicationJSON, e)
				if ceer != nil {
					a.logger.Errorf("failed to set event data: %v", err)
					return
				}

				if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
					a.logger.Errorf("failed to send event data: %v", err)
					return
				}

				a.logger.Debug("Sent Cloudevent Sucessfully")
			}
		}
	}

}

type event struct {
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
}
