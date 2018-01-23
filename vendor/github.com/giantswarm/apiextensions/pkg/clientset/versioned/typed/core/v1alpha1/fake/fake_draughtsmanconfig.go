/*
Copyright 2018 The Kubernetes Authors.

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

package fake

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDraughtsmanConfigs implements DraughtsmanConfigInterface
type FakeDraughtsmanConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var draughtsmanconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "draughtsmanconfigs"}

var draughtsmanconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "DraughtsmanConfig"}

// Get takes name of the draughtsmanConfig, and returns the corresponding draughtsmanConfig object, and an error if there is any.
func (c *FakeDraughtsmanConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.DraughtsmanConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(draughtsmanconfigsResource, c.ns, name), &v1alpha1.DraughtsmanConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DraughtsmanConfig), err
}

// List takes label and field selectors, and returns the list of DraughtsmanConfigs that match those selectors.
func (c *FakeDraughtsmanConfigs) List(opts v1.ListOptions) (result *v1alpha1.DraughtsmanConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(draughtsmanconfigsResource, draughtsmanconfigsKind, c.ns, opts), &v1alpha1.DraughtsmanConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DraughtsmanConfigList{}
	for _, item := range obj.(*v1alpha1.DraughtsmanConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested draughtsmanConfigs.
func (c *FakeDraughtsmanConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(draughtsmanconfigsResource, c.ns, opts))

}

// Create takes the representation of a draughtsmanConfig and creates it.  Returns the server's representation of the draughtsmanConfig, and an error, if there is any.
func (c *FakeDraughtsmanConfigs) Create(draughtsmanConfig *v1alpha1.DraughtsmanConfig) (result *v1alpha1.DraughtsmanConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(draughtsmanconfigsResource, c.ns, draughtsmanConfig), &v1alpha1.DraughtsmanConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DraughtsmanConfig), err
}

// Update takes the representation of a draughtsmanConfig and updates it. Returns the server's representation of the draughtsmanConfig, and an error, if there is any.
func (c *FakeDraughtsmanConfigs) Update(draughtsmanConfig *v1alpha1.DraughtsmanConfig) (result *v1alpha1.DraughtsmanConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(draughtsmanconfigsResource, c.ns, draughtsmanConfig), &v1alpha1.DraughtsmanConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DraughtsmanConfig), err
}

// Delete takes name of the draughtsmanConfig and deletes it. Returns an error if one occurs.
func (c *FakeDraughtsmanConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(draughtsmanconfigsResource, c.ns, name), &v1alpha1.DraughtsmanConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDraughtsmanConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(draughtsmanconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DraughtsmanConfigList{})
	return err
}

// Patch applies the patch and returns the patched draughtsmanConfig.
func (c *FakeDraughtsmanConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DraughtsmanConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(draughtsmanconfigsResource, c.ns, name, data, subresources...), &v1alpha1.DraughtsmanConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DraughtsmanConfig), err
}
