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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAWSSQSSources implements AWSSQSSourceInterface
type FakeAWSSQSSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var awssqssourcesResource = schema.GroupVersionResource{Group: "sources.triggermesh.com", Version: "v1alpha1", Resource: "awssqssources"}

var awssqssourcesKind = schema.GroupVersionKind{Group: "sources.triggermesh.com", Version: "v1alpha1", Kind: "AWSSQSSource"}

// Get takes name of the aWSSQSSource, and returns the corresponding aWSSQSSource object, and an error if there is any.
func (c *FakeAWSSQSSources) Get(name string, options v1.GetOptions) (result *v1alpha1.AWSSQSSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(awssqssourcesResource, c.ns, name), &v1alpha1.AWSSQSSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWSSQSSource), err
}

// List takes label and field selectors, and returns the list of AWSSQSSources that match those selectors.
func (c *FakeAWSSQSSources) List(opts v1.ListOptions) (result *v1alpha1.AWSSQSSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(awssqssourcesResource, awssqssourcesKind, c.ns, opts), &v1alpha1.AWSSQSSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AWSSQSSourceList{ListMeta: obj.(*v1alpha1.AWSSQSSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.AWSSQSSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested aWSSQSSources.
func (c *FakeAWSSQSSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(awssqssourcesResource, c.ns, opts))

}

// Create takes the representation of a aWSSQSSource and creates it.  Returns the server's representation of the aWSSQSSource, and an error, if there is any.
func (c *FakeAWSSQSSources) Create(aWSSQSSource *v1alpha1.AWSSQSSource) (result *v1alpha1.AWSSQSSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(awssqssourcesResource, c.ns, aWSSQSSource), &v1alpha1.AWSSQSSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWSSQSSource), err
}

// Update takes the representation of a aWSSQSSource and updates it. Returns the server's representation of the aWSSQSSource, and an error, if there is any.
func (c *FakeAWSSQSSources) Update(aWSSQSSource *v1alpha1.AWSSQSSource) (result *v1alpha1.AWSSQSSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(awssqssourcesResource, c.ns, aWSSQSSource), &v1alpha1.AWSSQSSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWSSQSSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAWSSQSSources) UpdateStatus(aWSSQSSource *v1alpha1.AWSSQSSource) (*v1alpha1.AWSSQSSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(awssqssourcesResource, "status", c.ns, aWSSQSSource), &v1alpha1.AWSSQSSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWSSQSSource), err
}

// Delete takes name of the aWSSQSSource and deletes it. Returns an error if one occurs.
func (c *FakeAWSSQSSources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(awssqssourcesResource, c.ns, name), &v1alpha1.AWSSQSSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAWSSQSSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(awssqssourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.AWSSQSSourceList{})
	return err
}

// Patch applies the patch and returns the patched aWSSQSSource.
func (c *FakeAWSSQSSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSSQSSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(awssqssourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.AWSSQSSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWSSQSSource), err
}