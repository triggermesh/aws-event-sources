/*
Copyright (c) 2020-2021 TriggerMesh Inc.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AWSCognitoIdentitySourceLister helps list AWSCognitoIdentitySources.
// All objects returned here must be treated as read-only.
type AWSCognitoIdentitySourceLister interface {
	// List lists all AWSCognitoIdentitySources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSCognitoIdentitySource, err error)
	// AWSCognitoIdentitySources returns an object that can list and get AWSCognitoIdentitySources.
	AWSCognitoIdentitySources(namespace string) AWSCognitoIdentitySourceNamespaceLister
	AWSCognitoIdentitySourceListerExpansion
}

// aWSCognitoIdentitySourceLister implements the AWSCognitoIdentitySourceLister interface.
type aWSCognitoIdentitySourceLister struct {
	indexer cache.Indexer
}

// NewAWSCognitoIdentitySourceLister returns a new AWSCognitoIdentitySourceLister.
func NewAWSCognitoIdentitySourceLister(indexer cache.Indexer) AWSCognitoIdentitySourceLister {
	return &aWSCognitoIdentitySourceLister{indexer: indexer}
}

// List lists all AWSCognitoIdentitySources in the indexer.
func (s *aWSCognitoIdentitySourceLister) List(selector labels.Selector) (ret []*v1alpha1.AWSCognitoIdentitySource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSCognitoIdentitySource))
	})
	return ret, err
}

// AWSCognitoIdentitySources returns an object that can list and get AWSCognitoIdentitySources.
func (s *aWSCognitoIdentitySourceLister) AWSCognitoIdentitySources(namespace string) AWSCognitoIdentitySourceNamespaceLister {
	return aWSCognitoIdentitySourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AWSCognitoIdentitySourceNamespaceLister helps list and get AWSCognitoIdentitySources.
// All objects returned here must be treated as read-only.
type AWSCognitoIdentitySourceNamespaceLister interface {
	// List lists all AWSCognitoIdentitySources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSCognitoIdentitySource, err error)
	// Get retrieves the AWSCognitoIdentitySource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.AWSCognitoIdentitySource, error)
	AWSCognitoIdentitySourceNamespaceListerExpansion
}

// aWSCognitoIdentitySourceNamespaceLister implements the AWSCognitoIdentitySourceNamespaceLister
// interface.
type aWSCognitoIdentitySourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AWSCognitoIdentitySources in the indexer for a given namespace.
func (s aWSCognitoIdentitySourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.AWSCognitoIdentitySource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSCognitoIdentitySource))
	})
	return ret, err
}

// Get retrieves the AWSCognitoIdentitySource from the indexer for a given namespace and name.
func (s aWSCognitoIdentitySourceNamespaceLister) Get(name string) (*v1alpha1.AWSCognitoIdentitySource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("awscognitoidentitysource"), name)
	}
	return obj.(*v1alpha1.AWSCognitoIdentitySource), nil
}
