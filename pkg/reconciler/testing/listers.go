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

package testing

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakek8sclient "k8s.io/client-go/kubernetes/fake"
	k8slistersv1 "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	rt "knative.dev/pkg/reconciler/testing"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	fakeclient "github.com/triggermesh/aws-event-sources/pkg/client/generated/clientset/internalclientset/fake"
	listersv1alpha1 "github.com/triggermesh/aws-event-sources/pkg/client/generated/listers/sources/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeclient.AddToScheme,
	fakek8sclient.AddToScheme,
	// although our reconcilers do not handle eventing objects directly, we
	// do need to register the eventing Scheme so that sink URI resolvers
	// can recongnize the Broker objects we use in tests
	fakeeventingclientset.AddToScheme,
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	sb := runtime.NewSchemeBuilder(clientSetSchemes...)
	if err := sb.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("error building Scheme: %s", err))
	}

	return scheme
}

type Listers struct {
	sorter rt.ObjectSorter
}

func NewListers(scheme *runtime.Scheme, objs []runtime.Object) Listers {
	ls := Listers{
		sorter: rt.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetSourcesObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeclient.AddToScheme)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakek8sclient.AddToScheme)
}

func (l *Listers) GetAWSCodeCommitSourceLister() listersv1alpha1.AWSCodeCommitSourceLister {
	return listersv1alpha1.NewAWSCodeCommitSourceLister(l.IndexerFor(&v1alpha1.AWSCodeCommitSource{}))
}

func (l *Listers) GetAWSCognitoSourceLister() listersv1alpha1.AWSCognitoSourceLister {
	return listersv1alpha1.NewAWSCognitoSourceLister(l.IndexerFor(&v1alpha1.AWSCognitoSource{}))
}

func (l *Listers) GetAWSDynamoDBSourceLister() listersv1alpha1.AWSDynamoDBSourceLister {
	return listersv1alpha1.NewAWSDynamoDBSourceLister(l.IndexerFor(&v1alpha1.AWSDynamoDBSource{}))
}

func (l *Listers) GetAWSIoTSourceLister() listersv1alpha1.AWSIoTSourceLister {
	return listersv1alpha1.NewAWSIoTSourceLister(l.IndexerFor(&v1alpha1.AWSIoTSource{}))
}

func (l *Listers) GetAWSKinesisSourceLister() listersv1alpha1.AWSKinesisSourceLister {
	return listersv1alpha1.NewAWSKinesisSourceLister(l.IndexerFor(&v1alpha1.AWSKinesisSource{}))
}

func (l *Listers) GetAWSSNSSourceLister() listersv1alpha1.AWSSNSSourceLister {
	return listersv1alpha1.NewAWSSNSSourceLister(l.IndexerFor(&v1alpha1.AWSSNSSource{}))
}

func (l *Listers) GetAWSSQSSourceLister() listersv1alpha1.AWSSQSSourceLister {
	return listersv1alpha1.NewAWSSQSSourceLister(l.IndexerFor(&v1alpha1.AWSSQSSource{}))
}

func (l *Listers) GetDeploymentLister() k8slistersv1.DeploymentLister {
	return k8slistersv1.NewDeploymentLister(l.IndexerFor(&appsv1.Deployment{}))
}
