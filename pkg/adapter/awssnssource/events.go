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
	"time"
)

// SNSEventRecord represents a SNS event.
type SNSEventRecord struct {
	EventVersion string    `json:"eventVersion"`
	EventSource  string    `json:"eventSource"`
	SNS          SNSEntity `json:"sns"`
}

// SNSEntity is the data from a SNS notification.
// see https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html#http-notification-json
type SNSEntity struct {
	Message           string                 `json:"message"`
	MessageID         string                 `json:"messageId"`
	Signature         string                 `json:"signature"`
	SignatureVersion  string                 `json:"signatureVersion"`
	SigningCertURL    string                 `json:"signingCertUrl"`
	Subject           string                 `json:"subject"`
	Timestamp         time.Time              `json:"timestamp"`
	TopicArn          string                 `json:"topicArn"`
	Type              string                 `json:"type"`
	UnsubscribeURL    string                 `json:"unsubscribeUrl"`
	MessageAttributes map[string]interface{} `json:"messageAttributes"`
}
