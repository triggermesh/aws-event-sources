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

package awscloudwatchsource

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/stretchr/testify/assert"
	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	loggingtesting "knative.dev/pkg/logging/testing"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

type mockCloudWatchClient struct {
	cloudwatchiface.CloudWatchAPI

	Resp cloudwatch.GetMetricDataOutput
	err  error
}

func (m mockCloudWatchClient) GetMetricDataPages(input *cloudwatch.GetMetricDataInput, fn func(*cloudwatch.GetMetricDataOutput, bool) bool) error {
	fn(&m.Resp, true)

	return m.err
}

// TestParseQueries Given a query string, ensure that
func TestParseQueries(t *testing.T) {
	queryStr := "[{\"name\":\"testquery\",\"metric\":{\"period\":60,\"stat\":\"Sum\",\"metric\":{\"dimensions\":[{\"name\":\"FunctionName\",\"value\":\"makemoney\"}],\"metricName\":\"Duration\",\"namespace\":\"AWS/Lambda\"}}}]"

	name := "testquery"
	period := int64(60)
	stat := "Sum"
	metricName := "Duration"
	metricNamespace := "AWS/Lambda"
	dimensionName := "FunctionName"
	dimensionValue := "makemoney"

	metricQuery := cloudwatch.MetricDataQuery{
		Expression: nil,
		Id:         &name,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Dimensions: []*cloudwatch.Dimension{{
					Name:  &dimensionName,
					Value: &dimensionValue,
				}},
				MetricName: &metricName,
				Namespace:  &metricNamespace,
			},
			Period: &period,
			Stat:   &stat,
		},
	}

	results, err := parseQueries(queryStr)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	m := results[0]
	assert.EqualValues(t, metricQuery, *m)
}

func TestCollectMetrics(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	metricId := "testmetrics"
	metricNamespace := "AWS/Lambda"
	metricLabel := "Duration"
	metricStatusCode := "Complete"
	ts, _ := time.Parse(time.RFC3339, time.Now().String())
	val := float64(37.566818845509246)
	pollingInterval, _ := time.ParseDuration("1m")

	dimensionName := "FunctionName"
	dimensionValue := "makemoney"
	period := int64(60)
	stat := "Sum"

	a := &adapter{
		logger:          loggingtesting.TestLogger(t),
		pollingInterval: pollingInterval,
		ceClient:        ceClient,
		metricQueries: []*cloudwatch.MetricDataQuery{{
			Expression: nil,
			Id:         &metricId,
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{{
						Name:  &dimensionName,
						Value: &dimensionValue,
					}},
					MetricName: &metricLabel,
					Namespace:  &metricNamespace,
				},
				Period: &period,
				Stat:   &stat,
			},
		}},
		cwClient: mockCloudWatchClient{
			Resp: cloudwatch.GetMetricDataOutput{
				Messages: nil,
				MetricDataResults: []*cloudwatch.MetricDataResult{{
					Id:         &metricId,
					Label:      &metricLabel,
					Messages:   nil,
					StatusCode: &metricStatusCode,
					Timestamps: []*time.Time{&ts},
					Values:     []*float64{&val},
				}},
				NextToken: nil,
			},
			err: nil,
		},
	}

	metricOutput := cloudwatch.GetMetricDataOutput{
		Messages: nil,
		MetricDataResults: []*cloudwatch.MetricDataResult{{
			Id:         &metricId,
			Label:      &metricLabel,
			Messages:   nil,
			StatusCode: &metricStatusCode,
			Timestamps: []*time.Time{&ts},
			Values:     []*float64{&val},
		}},
		NextToken: nil,
	}

	a.CollectMetrics(time.Now())

	events := ceClient.Sent()
	assert.Len(t, events, 1)

	assert.EqualValues(t, events[0].Type(), v1alpha1.AWSEventType(metricEventType, "metric"))
	assert.EqualValues(t, events[0].Source(), "testmetrics")

	var metricRecord cloudwatch.MetricDataResult
	err := events[0].DataAs(&metricRecord)

	assert.NoError(t, err)
	assert.EqualValues(t, *metricOutput.MetricDataResults[0], metricRecord)
}

func TestSendMetricEvent(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	metricId := "testmetrics"
	metricLabel := "Duration"
	metricStatusCode := "Complete"
	ts, _ := time.Parse(time.RFC3339, time.Now().String())
	val := float64(37.566818845509246) // must keep this cast to ensure proper [de]serialization

	metricOutput := cloudwatch.GetMetricDataOutput{
		Messages: nil,
		MetricDataResults: []*cloudwatch.MetricDataResult{{
			Id:         &metricId,
			Label:      &metricLabel,
			Messages:   nil,
			StatusCode: &metricStatusCode,
			Timestamps: []*time.Time{&ts},
			Values:     []*float64{&val},
		}},
		NextToken: nil,
	}

	err := a.SendMetricEvent(&metricOutput, "testname")
	assert.NoError(t, err)
	events := ceClient.Sent()
	assert.Len(t, events, 1)

	assert.EqualValues(t, events[0].Type(), v1alpha1.AWSEventType(metricEventType, "metric"))
	assert.EqualValues(t, events[0].Source(), "testmetrics")

	var metricRecord cloudwatch.MetricDataResult
	err = events[0].DataAs(&metricRecord)

	assert.NoError(t, err)
	assert.EqualValues(t, *metricOutput.MetricDataResults[0], metricRecord)
}

func TestSendMessageEvent(t *testing.T) {
	ceClient := adaptertest.NewTestClient()

	a := &adapter{
		logger:   loggingtesting.TestLogger(t),
		ceClient: ceClient,
	}

	msgCode := "Success"
	msgValue := "This is a sample message value"

	metricOutput := cloudwatch.GetMetricDataOutput{
		Messages: []*cloudwatch.MessageData{{
			Code:  &msgCode,
			Value: &msgValue,
		}},
		NextToken: nil,
	}

	err := a.SendMetricEvent(&metricOutput, "testname")
	assert.NoError(t, err)
	events := ceClient.Sent()
	assert.Len(t, events, 1)

	assert.EqualValues(t, events[0].Type(), v1alpha1.AWSEventType(metricEventType, "message"))
	assert.EqualValues(t, events[0].Source(), "testname-0")

	var metricRecord cloudwatch.MessageData
	err = events[0].DataAs(&metricRecord)

	assert.NoError(t, err)
	assert.EqualValues(t, *metricOutput.Messages[0], metricRecord)
}
