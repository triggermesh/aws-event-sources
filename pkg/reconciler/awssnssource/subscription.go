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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// ensureSubscribed ensures the source's HTTP(S) endpoint is subscribed to the
// SNS topic.
func (r *Reconciler) ensureSubscribed(ctx context.Context) error {
	src := v1alpha1.SourceFromContext(ctx)

	adapter, err := r.base.FindAdapter(src)
	switch {
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return fmt.Errorf("finding receive adapter: %w", err)
	}

	url := adapter.Status.URL

	// skip this cycle if the adapter URL wasn't yet determined
	if !adapter.IsReady() || url == nil {
		return nil
	}

	spec := src.(apis.HasSpec).GetUntypedSpec().(v1alpha1.AWSSNSSourceSpec)

	snsClient, err := newSNSClient(r.secretsCli(src.GetNamespace()), spec.ARN.Region, &spec.Credentials)
	if err != nil {
		return fmt.Errorf("creating SNS client: %w", err)
	}

	resp, err := snsClient.SubscribeWithContext(ctx, &sns.SubscribeInput{
		Endpoint:              aws.String(url.String()),
		Protocol:              &url.Scheme,
		TopicArn:              aws.String(spec.ARN.String()),
		Attributes:            spec.SubscriptionAttributes,
		ReturnSubscriptionArn: aws.Bool(true),
	})

	switch {
	case isDenied(err):
		return controller.NewPermanentError(fmt.Errorf("subscribing to SNS topic: %w", err))

	case err != nil:
		return fmt.Errorf("subscribing to SNS topic: %w", err)
	}

	logging.FromContext(ctx).Debug("Subscribe responded with: ", resp)

	return nil
}

// newSNSClient returns a new SNS client for the given region using static credentials.
func newSNSClient(cli coreclientv1.SecretInterface,
	region string, creds *v1alpha1.AWSSecurityCredentials) (*sns.SNS, error) {

	credsValue, err := awsCredentials(cli, creds)
	if err != nil {
		return nil, fmt.Errorf("reading AWS security credentials: %w", err)
	}

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(region).
		WithCredentials(credentials.NewStaticCredentialsFromCreds(*credsValue)),
	))

	return sns.New(cfg), nil
}

// awsCredentials returns the AWS security credentials referenced in the
// source's spec.
func awsCredentials(cli coreclientv1.SecretInterface,
	creds *v1alpha1.AWSSecurityCredentials) (*credentials.Value, error) {

	accessKeyID := creds.AccessKeyID.Value
	secretAccessKey := creds.SecretAccessKey.Value

	// cache a Secret object by name to avoid GET-ing the same Secret
	// object multiple times
	var secretCache map[string]*corev1.Secret

	if vfs := creds.AccessKeyID.ValueFromSecret; vfs != nil {
		secr, err := cli.Get(vfs.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// cache Secret containing the access key ID so it can be reused
		// below in case the same Secret contains the secret access key
		secretCache = map[string]*corev1.Secret{
			vfs.Name: secr,
		}

		accessKeyID = string(secr.Data[vfs.Key])
	}

	if vfs := creds.SecretAccessKey.ValueFromSecret; vfs != nil {
		var secr *corev1.Secret
		var err error

		if secretCache != nil && secretCache[vfs.Name] != nil {
			secr = secretCache[vfs.Name]
		} else {
			secr, err = cli.Get(vfs.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		secretAccessKey = string(secr.Data[vfs.Key])
	}

	return &credentials.Value{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}

// isDenied returns whether the given error indicates that a request to the SNS
// API could not be authorized.
func isDenied(err error) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == sns.ErrCodeAuthorizationErrorException
	}
	return false
}
