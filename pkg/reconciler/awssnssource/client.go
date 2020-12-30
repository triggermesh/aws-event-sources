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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
)

// Client is an alias for the SNSAPI interface.
type Client = snsiface.SNSAPI

// ClientGetter can obtain SNS clients.
type ClientGetter interface {
	Get(*v1alpha1.AWSSNSSource) (Client, error)
}

// newClientGetter returns a ClientGetter for the given namespacedSecretsGetter.
func newClientGetter(sg namespacedSecretsGetter) *clientGetterWithSecretGetter {
	return &clientGetterWithSecretGetter{
		sg: sg,
	}
}

type namespacedSecretsGetter func(namespace string) coreclientv1.SecretInterface

// clientGetterWithSecretGetter gets SNS clients using static credentials
// retrieved using a Secret getter.
type clientGetterWithSecretGetter struct {
	sg namespacedSecretsGetter
}

// clientGetterWithSecretGetter implements ClientGetter.
var _ ClientGetter = (*clientGetterWithSecretGetter)(nil)

// Get implements ClientGetter.
func (g *clientGetterWithSecretGetter) Get(src *v1alpha1.AWSSNSSource) (Client, error) {
	creds, err := g.awsCredentials(src.Namespace, &src.Spec.Credentials)
	if err != nil {
		return nil, fmt.Errorf("retrieving AWS security credentials: %w", err)
	}

	return sns.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion(src.Spec.ARN.Region).
		WithCredentials(credentials.NewStaticCredentialsFromCreds(*creds)),
	))), nil
}

// awsCredentials returns the AWS security credentials referenced in a source's
// spec, using the ClientGetter's Secrets getter if necessary.
func (g *clientGetterWithSecretGetter) awsCredentials(namespace string,
	creds *v1alpha1.AWSSecurityCredentials) (*credentials.Value, error) {

	accessKeyID := creds.AccessKeyID.Value
	secretAccessKey := creds.SecretAccessKey.Value

	cli := g.sg(namespace)

	// cache a Secret object by name to avoid GET-ing the same Secret
	// object multiple times
	var secretCache map[string]*corev1.Secret

	if vfs := creds.AccessKeyID.ValueFromSecret; vfs != nil {
		secr, err := cli.Get(context.Background(), vfs.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("getting Secret from cluster: %w", err)
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
			secr, err = cli.Get(context.Background(), vfs.Name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("getting Secret from cluster: %w", err)
			}
		}

		secretAccessKey = string(secr.Data[vfs.Key])
	}

	return &credentials.Value{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}

// clientGetterFunc allows the use of ordinary functions as ClientGetter.
type clientGetterFunc func(*v1alpha1.AWSSNSSource) (Client, error)

// clientGetterFunc implements ClientGetter.
var _ ClientGetter = (clientGetterFunc)(nil)

// Get implements ClientGetter.
func (f clientGetterFunc) Get(src *v1alpha1.AWSSNSSource) (Client, error) {
	return f(src)
}
