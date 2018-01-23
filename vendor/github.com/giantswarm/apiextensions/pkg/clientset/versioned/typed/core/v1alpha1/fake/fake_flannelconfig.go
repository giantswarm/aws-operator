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

// FakeFlannelConfigs implements FlannelConfigInterface
type FakeFlannelConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var flannelconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "flannelconfigs"}

var flannelconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "FlannelConfig"}

// Get takes name of the flannelConfig, and returns the corresponding flannelConfig object, and an error if there is any.
func (c *FakeFlannelConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.FlannelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(flannelconfigsResource, c.ns, name), &v1alpha1.FlannelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FlannelConfig), err
}

// List takes label and field selectors, and returns the list of FlannelConfigs that match those selectors.
func (c *FakeFlannelConfigs) List(opts v1.ListOptions) (result *v1alpha1.FlannelConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(flannelconfigsResource, flannelconfigsKind, c.ns, opts), &v1alpha1.FlannelConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FlannelConfigList{}
	for _, item := range obj.(*v1alpha1.FlannelConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested flannelConfigs.
func (c *FakeFlannelConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(flannelconfigsResource, c.ns, opts))

}

// Create takes the representation of a flannelConfig and creates it.  Returns the server's representation of the flannelConfig, and an error, if there is any.
func (c *FakeFlannelConfigs) Create(flannelConfig *v1alpha1.FlannelConfig) (result *v1alpha1.FlannelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(flannelconfigsResource, c.ns, flannelConfig), &v1alpha1.FlannelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FlannelConfig), err
}

// Update takes the representation of a flannelConfig and updates it. Returns the server's representation of the flannelConfig, and an error, if there is any.
func (c *FakeFlannelConfigs) Update(flannelConfig *v1alpha1.FlannelConfig) (result *v1alpha1.FlannelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(flannelconfigsResource, c.ns, flannelConfig), &v1alpha1.FlannelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FlannelConfig), err
}

// Delete takes name of the flannelConfig and deletes it. Returns an error if one occurs.
func (c *FakeFlannelConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(flannelconfigsResource, c.ns, name), &v1alpha1.FlannelConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFlannelConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(flannelconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.FlannelConfigList{})
	return err
}

// Patch applies the patch and returns the patched flannelConfig.
func (c *FakeFlannelConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.FlannelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(flannelconfigsResource, c.ns, name, data, subresources...), &v1alpha1.FlannelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FlannelConfig), err
}
