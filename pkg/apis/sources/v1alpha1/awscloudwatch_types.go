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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSCloudWatchSource is the Schema for the event source.
type AWSCloudWatchSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSCloudWatchSourceSpec `json:"spec,omitempty"`
	Status EventSourceStatus       `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object = (*AWSCloudWatchSource)(nil)
	_ EventSource    = (*AWSCloudWatchSource)(nil)
)

// AWSCloudWatchSourceSpec defines the desired state of the event source.
type AWSCloudWatchSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	// AWS Region for metrics
	Region string `json:"region"`
	// List of metric queries
	// +optional
	MetricQueries *[]AWSCloudWatchMetricQueries `json:"metricQueries,omitempty"`
	// PollingFrequency in a duration format for how often to pull metrics data from. Default is 5m
	// +optional
	PollingFrequency *string `json:"pollingFrequency,omitempty"`

	// Credentials to interact with the AWS CodeCommit API.
	Credentials AWSSecurityCredentials `json:"credentials"`
}

// Define the metric to return. Consult the AWS CloudWatch API Guide for details:
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/Welcome.html
type AWSCloudWatchMetricQueries struct {
	// Unique short-name identify the query
	Name string `json:"name"`
	// Math expression for calculating metrics. Can have this or a metric
	// +optional
	Expression *string `json:"expression,omitempty"`
	// Metric for retrieving specific metrics
	// +optional
	Metric *AWSCloudWatchMetricStat `json:"metric,omitempty"`
}

type AWSCloudWatchMetricStat struct {
	Metric AWSCloudWatchMetric `json:"metric"`         // Definition of the metric
	Period int64               `json:"period"`         // metric resolution in seconds
	Stat   string              `json:"stat"`           // statistic type to use
	Unit   string              `json:"unit,omitempty"` // The unit of the metric being returned
}

type AWSCloudWatchMetric struct {
	Dimensions []AWSCloudWatchMetricDimension `json:"dimensions"`
	MetricName string                         `json:"metricName"`
	Namespace  string                         `json:"namespace"`
}

type AWSCloudWatchMetricDimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSCloudWatchSourceList contains a list of event sources.
type AWSCloudWatchSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSCloudWatchSource `json:"items"`
}
