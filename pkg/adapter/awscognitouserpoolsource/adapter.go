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

package awscognitouserpoolsource

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

var (
	// number of seconds to wait between poll requests
	// TODO: replace with exponential backoff
	pollInterval = 5
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

	cgnIdentityClient cognitoidentityprovideriface.CognitoIdentityProviderAPI
	ceClient          cloudevents.Client

	arn        arn.ARN
	userPoolID string
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

		cgnIdentityClient: cognitoidentityprovider.New(cfg),
		ceClient:          ceClient,

		arn:        arn,
		userPoolID: common.MustParseCognitoUserPoolID(arn.Resource),
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(stopCh <-chan struct{}) error {
	a.logger.Infof("Listening to AWS Cognito User Pool: %s", a.userPoolID)

	var latestTimestamp time.Time

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			a.logger.Info("Exiting receiver loop")
			return nil
		case <-ticker.C:
			users, err := a.listUsers()
			if err != nil {
				a.logger.Errorf("Cognito ListUsers failed: %v", err)
				continue
			}

			users, latestTimestamp = filterByTimestamp(users, latestTimestamp)

			for _, user := range users {
				err := a.sendCognitoEvent(user)
				if err != nil {
					a.logger.Errorf("Failed to send cloudevent: %v", err)
				}
			}
		}
	}
}

func (a *adapter) listUsers() ([]*cognitoidentityprovider.UserType, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: &a.userPoolID,
	}
	output, err := a.cgnIdentityClient.ListUsers(input)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, fmt.Errorf("cognito ListUsers response is nil")
	}
	return output.Users, nil
}

func filterByTimestamp(users []*cognitoidentityprovider.UserType, latestTimestamp time.Time) ([]*cognitoidentityprovider.UserType, time.Time) {
	var newUsers []*cognitoidentityprovider.UserType
	newLatestTimestamp := latestTimestamp
	for _, user := range users {
		// Get latest modification timestamp from users list
		// and store it in temporary variable
		if user.UserLastModifiedDate.After(newLatestTimestamp) {
			newLatestTimestamp = *user.UserLastModifiedDate
		}
		// latest.isZero() true in first iteration - do not send already existing users.
		// Also, do not send user object if it was not modified after our latest timestamp mark.
		// (we use "not after" because "before" will be always false for the last created user)
		if latestTimestamp.IsZero() || !user.UserLastModifiedDate.After(latestTimestamp) {
			continue
		}
		newUsers = append(newUsers, user)
	}
	return newUsers, newLatestTimestamp
}

func (a *adapter) sendCognitoEvent(user *cognitoidentityprovider.UserType) error {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetSubject(a.userPoolID)
	event.SetSource(a.arn.String())
	event.SetID(string(uuid.NewUUID()))
	event.SetType(v1alpha1.AWSEventType(a.arn.Service, v1alpha1.AWSCognitoGenericEventType))
	if err := event.SetData(cloudevents.ApplicationJSON, user); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	if result := a.ceClient.Send(context.Background(), event); !cloudevents.IsACK(result) {
		return result
	}
	return nil
}
