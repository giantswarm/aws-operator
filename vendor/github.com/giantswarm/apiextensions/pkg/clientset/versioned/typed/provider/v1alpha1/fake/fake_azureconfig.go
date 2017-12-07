/*
Copyright 2017 The Kubernetes Authors.

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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAzureConfigs implements AzureConfigInterface
type FakeAzureConfigs struct {
	Fake *FakeProviderV1alpha1
	ns   string
}

var azureconfigsResource = schema.GroupVersionResource{Group: "provider.giantswarm.io", Version: "v1alpha1", Resource: "azureconfigs"}

var azureconfigsKind = schema.GroupVersionKind{Group: "provider.giantswarm.io", Version: "v1alpha1", Kind: "AzureConfig"}

// Get takes name of the azureConfig, and returns the corresponding azureConfig object, and an error if there is any.
func (c *FakeAzureConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.AzureConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(azureconfigsResource, c.ns, name), &v1alpha1.AzureConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AzureConfig), err
}

// List takes label and field selectors, and returns the list of AzureConfigs that match those selectors.
func (c *FakeAzureConfigs) List(opts v1.ListOptions) (result *v1alpha1.AzureConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(azureconfigsResource, azureconfigsKind, c.ns, opts), &v1alpha1.AzureConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AzureConfigList{}
	for _, item := range obj.(*v1alpha1.AzureConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested azureConfigs.
func (c *FakeAzureConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(azureconfigsResource, c.ns, opts))

}

// Create takes the representation of a azureConfig and creates it.  Returns the server's representation of the azureConfig, and an error, if there is any.
func (c *FakeAzureConfigs) Create(azureConfig *v1alpha1.AzureConfig) (result *v1alpha1.AzureConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(azureconfigsResource, c.ns, azureConfig), &v1alpha1.AzureConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AzureConfig), err
}

// Update takes the representation of a azureConfig and updates it. Returns the server's representation of the azureConfig, and an error, if there is any.
func (c *FakeAzureConfigs) Update(azureConfig *v1alpha1.AzureConfig) (result *v1alpha1.AzureConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(azureconfigsResource, c.ns, azureConfig), &v1alpha1.AzureConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AzureConfig), err
}

// Delete takes name of the azureConfig and deletes it. Returns an error if one occurs.
func (c *FakeAzureConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(azureconfigsResource, c.ns, name), &v1alpha1.AzureConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAzureConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(azureconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.AzureConfigList{})
	return err
}

// Patch applies the patch and returns the patched azureConfig.
func (c *FakeAzureConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AzureConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(azureconfigsResource, c.ns, name, data, subresources...), &v1alpha1.AzureConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AzureConfig), err
}
