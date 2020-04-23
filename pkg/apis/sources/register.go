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

package sources

import "k8s.io/apimachinery/pkg/runtime/schema"

// GroupName is the name of the API group this package's resources belong to.
const GroupName = "sources.triggermesh.io"

var (
	// AWSCodeCommitSourceResource respresents an event source for AWS CodeCommit.
	AWSCodeCommitSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awscodecommitsources",
	}

	// AWSCognitoSourceResource respresents an event source for AWS Cognito.
	AWSCognitoSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awscognitosources",
	}

	// AWSDynamoDBSourceResource respresents an event source for AWS DynamoDB.
	AWSDynamoDBSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awsdynamodbsources",
	}

	// AWSIoTSourceResource respresents an event source for AWS IoT.
	AWSIoTSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awsiotsources",
	}

	// AWSKinesisSourceResource respresents an event source for AWS Kinesis.
	AWSKinesisSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awskinesissources",
	}

	// AWSSNSSourceResource respresents an event source for AWS SNS.
	AWSSNSSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awssnssources",
	}

	// AWSSQSSourceResource respresents an event source for AWS SQS.
	AWSSQSSourceResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "awssqssources",
	}
)
