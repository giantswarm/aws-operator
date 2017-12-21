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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeStorageConfigs implements StorageConfigInterface
type FakeStorageConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var storageconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "storageconfigs"}

var storageconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "StorageConfig"}

// Get takes name of the storageConfig, and returns the corresponding storageConfig object, and an error if there is any.
func (c *FakeStorageConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.StorageConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(storageconfigsResource, c.ns, name), &v1alpha1.StorageConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageConfig), err
}

// List takes label and field selectors, and returns the list of StorageConfigs that match those selectors.
func (c *FakeStorageConfigs) List(opts v1.ListOptions) (result *v1alpha1.StorageConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(storageconfigsResource, storageconfigsKind, c.ns, opts), &v1alpha1.StorageConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.StorageConfigList{}
	for _, item := range obj.(*v1alpha1.StorageConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storageConfigs.
func (c *FakeStorageConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(storageconfigsResource, c.ns, opts))

}

// Create takes the representation of a storageConfig and creates it.  Returns the server's representation of the storageConfig, and an error, if there is any.
func (c *FakeStorageConfigs) Create(storageConfig *v1alpha1.StorageConfig) (result *v1alpha1.StorageConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(storageconfigsResource, c.ns, storageConfig), &v1alpha1.StorageConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageConfig), err
}

// Update takes the representation of a storageConfig and updates it. Returns the server's representation of the storageConfig, and an error, if there is any.
func (c *FakeStorageConfigs) Update(storageConfig *v1alpha1.StorageConfig) (result *v1alpha1.StorageConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(storageconfigsResource, c.ns, storageConfig), &v1alpha1.StorageConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageConfig), err
}

// Delete takes name of the storageConfig and deletes it. Returns an error if one occurs.
func (c *FakeStorageConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(storageconfigsResource, c.ns, name), &v1alpha1.StorageConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStorageConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(storageconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.StorageConfigList{})
	return err
}

// Patch applies the patch and returns the patched storageConfig.
func (c *FakeStorageConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StorageConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(storageconfigsResource, c.ns, name, data, subresources...), &v1alpha1.StorageConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageConfig), err
}
