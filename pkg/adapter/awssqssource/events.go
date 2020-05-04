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

package awssqssource

import (
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Event represents the payload of an AWS SQS Event.
type Event struct {
	MessageID         *string                               `json:"messageId"`
	ReceiptHandle     *string                               `json:"receiptHandle"`
	Body              *string                               `json:"body"`
	Attributes        map[string]*string                    `json:"attributes"`
	MessageAttributes map[string]*sqs.MessageAttributeValue `json:"messageAttributes"`
	Md5OfBody         *string                               `json:"md5OfBody"`
	EventSource       *string                               `json:"eventSource"`
	EventSourceARN    *string                               `json:"eventSourceARN"`
	AwsRegion         *string                               `json:"awsRegion"`
}
