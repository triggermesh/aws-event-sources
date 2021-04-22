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

// AWSIoTSourceLister helps list AWSIoTSources.
// All objects returned here must be treated as read-only.
type AWSIoTSourceLister interface {
	// List lists all AWSIoTSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSIoTSource, err error)
	// AWSIoTSources returns an object that can list and get AWSIoTSources.
	AWSIoTSources(namespace string) AWSIoTSourceNamespaceLister
	AWSIoTSourceListerExpansion
}

// aWSIoTSourceLister implements the AWSIoTSourceLister interface.
type aWSIoTSourceLister struct {
	indexer cache.Indexer
}

// NewAWSIoTSourceLister returns a new AWSIoTSourceLister.
func NewAWSIoTSourceLister(indexer cache.Indexer) AWSIoTSourceLister {
	return &aWSIoTSourceLister{indexer: indexer}
}

// List lists all AWSIoTSources in the indexer.
func (s *aWSIoTSourceLister) List(selector labels.Selector) (ret []*v1alpha1.AWSIoTSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSIoTSource))
	})
	return ret, err
}

// AWSIoTSources returns an object that can list and get AWSIoTSources.
func (s *aWSIoTSourceLister) AWSIoTSources(namespace string) AWSIoTSourceNamespaceLister {
	return aWSIoTSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AWSIoTSourceNamespaceLister helps list and get AWSIoTSources.
// All objects returned here must be treated as read-only.
type AWSIoTSourceNamespaceLister interface {
	// List lists all AWSIoTSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSIoTSource, err error)
	// Get retrieves the AWSIoTSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.AWSIoTSource, error)
	AWSIoTSourceNamespaceListerExpansion
}

// aWSIoTSourceNamespaceLister implements the AWSIoTSourceNamespaceLister
// interface.
type aWSIoTSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AWSIoTSources in the indexer for a given namespace.
func (s aWSIoTSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.AWSIoTSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSIoTSource))
	})
	return ret, err
}

// Get retrieves the AWSIoTSource from the indexer for a given namespace and name.
func (s aWSIoTSourceNamespaceLister) Get(name string) (*v1alpha1.AWSIoTSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("awsiotsource"), name)
	}
	return obj.(*v1alpha1.AWSIoTSource), nil
}
