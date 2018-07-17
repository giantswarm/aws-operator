/*
Copyright 2018 Giant Swarm GmbH.

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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDrainerConfigs implements DrainerConfigInterface
type FakeDrainerConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var drainerconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "drainerconfigs"}

var drainerconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "DrainerConfig"}

// Get takes name of the drainerConfig, and returns the corresponding drainerConfig object, and an error if there is any.
func (c *FakeDrainerConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.DrainerConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(drainerconfigsResource, c.ns, name), &v1alpha1.DrainerConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DrainerConfig), err
}

// List takes label and field selectors, and returns the list of DrainerConfigs that match those selectors.
func (c *FakeDrainerConfigs) List(opts v1.ListOptions) (result *v1alpha1.DrainerConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(drainerconfigsResource, drainerconfigsKind, c.ns, opts), &v1alpha1.DrainerConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DrainerConfigList{}
	for _, item := range obj.(*v1alpha1.DrainerConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested drainerConfigs.
func (c *FakeDrainerConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(drainerconfigsResource, c.ns, opts))

}

// Create takes the representation of a drainerConfig and creates it.  Returns the server's representation of the drainerConfig, and an error, if there is any.
func (c *FakeDrainerConfigs) Create(drainerConfig *v1alpha1.DrainerConfig) (result *v1alpha1.DrainerConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(drainerconfigsResource, c.ns, drainerConfig), &v1alpha1.DrainerConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DrainerConfig), err
}

// Update takes the representation of a drainerConfig and updates it. Returns the server's representation of the drainerConfig, and an error, if there is any.
func (c *FakeDrainerConfigs) Update(drainerConfig *v1alpha1.DrainerConfig) (result *v1alpha1.DrainerConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(drainerconfigsResource, c.ns, drainerConfig), &v1alpha1.DrainerConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DrainerConfig), err
}

// Delete takes name of the drainerConfig and deletes it. Returns an error if one occurs.
func (c *FakeDrainerConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(drainerconfigsResource, c.ns, name), &v1alpha1.DrainerConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDrainerConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(drainerconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DrainerConfigList{})
	return err
}

// Patch applies the patch and returns the patched drainerConfig.
func (c *FakeDrainerConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DrainerConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(drainerconfigsResource, c.ns, name, data, subresources...), &v1alpha1.DrainerConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DrainerConfig), err
}
