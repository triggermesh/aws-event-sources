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

package awscodecommitsource

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/kubernetes/test/e2e/framework"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"

	"github.com/triggermesh/aws-event-sources/pkg/apis"
	apisources "github.com/triggermesh/aws-event-sources/pkg/apis/sources"
	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/test/e2e/framework/awscodecommit"
	"github.com/triggermesh/aws-event-sources/test/e2e/framework/conversion"
	"github.com/triggermesh/aws-event-sources/test/e2e/framework/sources"
)

var _ = Describe("AWS CodeCommit source", func() {
	var (
		srcClient dynamic.ResourceInterface
		ccClient  codecommitiface.CodeCommitAPI

		sink     duckv1.Destination
		repoARN  apis.ARN
		awsCreds credentials.Value
	)

	f := framework.NewDefaultFramework("awscodecommitsource")

	BeforeEach(func() {
		srcClient = f.DynamicClient.
			Resource(conversion.V1alpha1GVR(apisources.AWSCodeCommitSourceResource)).
			Namespace(f.Namespace.Name)

		By("reading AWS credentials", func() {
			sess := session.Must(session.NewSession())
			ccClient = codecommit.New(sess)
			awsCreds = readAWSCredentials(sess)
		})

		By("creating a Git repository", func() {
			arn, undoCreateRepo := awscodecommit.CreateRepository(ccClient, f)
			framework.AddCleanupAction(
				undoCreateRepo,
			)
			repoARN = parseARN(arn)
		})

		By("creating an event sink", func() {
			sink = sources.CreateEventDisplaySink(f.ClientSet, f.Namespace.Name)
		})
	})

	It("should generate events on selected actions", func() {
		By("creating an AWSCodeCommitSource object", func() {
			src, err := createSource(srcClient, f.Namespace.Name, "test-", sink,
				withARN(repoARN),
				withCredentials(awsCreds),
			)
			Expect(err).ToNot(HaveOccurred())

			sources.WaitForResourceReady(f.DynamicClient, src)
		})

		By("creating a Git commit", func() {
			awscodecommit.CreateCommit(ccClient, repoARN.Resource)
		})

		// TODO(antoineco): check event-display logs, metrics?
	})

	It("should reject invalid specs", func() {
		By("creating an AWSCodeCommitSource object with an invalid ARN", func() {
			invalidARN := repoARN
			invalidARN.AccountID = "invalid"

			_, err := createSource(srcClient, f.Namespace.Name, "test-invalid-arn", sink,
				withARN(invalidARN),
				withCredentials(awsCreds),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.arn: Invalid value: "))
		})

		By("creating an AWSCodeCommitSource object without credentials", func() {
			_, err := createSource(srcClient, f.Namespace.Name, "test-nocreds", sink,
				withARN(repoARN),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				`"spec.credentials.accessKeyID" must validate one and only one schema (oneOf).`))
		})

		By("creating an AWSCodeCommitSource object with invalid event types", func() {
			_, err := createSource(srcClient, f.Namespace.Name, "test-invalid-eventtypes", sink,
				withARN(repoARN),
				withEventType("invalid"),
				withCredentials(awsCreds),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`spec.eventTypes: Unsupported value: "invalid"`))
		})
	})
})

// createSource creates an AWSCodeCommitSource object initialized with the given options.
func createSource(srcClient dynamic.ResourceInterface, namespace, namePrefix string,
	sink duckv1.Destination, opts ...sourceOption) (*unstructured.Unstructured, error) {

	src := &v1alpha1.AWSCodeCommitSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: namePrefix,
		},
		Spec: v1alpha1.AWSCodeCommitSourceSpec{
			Branch: awscodecommit.DefaultBranch,
			SourceSpec: duckv1.SourceSpec{
				Sink: sink,
			},
		},
	}

	for _, opt := range opts {
		opt(src)
	}

	// set some sane defaults
	if src.Spec.EventTypes == nil {
		src.Spec.EventTypes = []string{"push"}
	}

	srcUnstr := conversion.ToUnstructured(src)

	return srcClient.Create(srcUnstr, metav1.CreateOptions{})
}

type sourceOption func(*v1alpha1.AWSCodeCommitSource)

func withARN(arn apis.ARN) sourceOption {
	return func(src *v1alpha1.AWSCodeCommitSource) {
		src.Spec.ARN = arn
	}
}

func withCredentials(creds credentials.Value) sourceOption {
	return func(src *v1alpha1.AWSCodeCommitSource) {
		src.Spec.Credentials = v1alpha1.AWSSecurityCredentials{
			AccessKeyID: v1alpha1.ValueFromField{
				Value: creds.AccessKeyID,
			},
			SecretAccessKey: v1alpha1.ValueFromField{
				Value: creds.SecretAccessKey,
			},
		}
	}
}

func withEventType(typ string) sourceOption {
	return func(src *v1alpha1.AWSCodeCommitSource) {
		src.Spec.EventTypes = append(src.Spec.EventTypes, typ)
	}
}

func parseARN(arnStr string) apis.ARN {
	arn, err := arn.Parse(arnStr)
	if err != nil {
		framework.Failf("Error parsing ARN string %q: %s", arnStr, err)
	}

	return apis.ARN(arn)
}

func readAWSCredentials(sess *session.Session) credentials.Value {
	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		framework.Failf("Error reading AWS credentials: %s", err)
	}

	return creds
}
