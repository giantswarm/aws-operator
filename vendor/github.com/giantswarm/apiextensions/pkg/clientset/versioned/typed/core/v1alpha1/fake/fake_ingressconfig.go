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

// FakeIngressConfigs implements IngressConfigInterface
type FakeIngressConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var ingressconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "ingressconfigs"}

var ingressconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "IngressConfig"}

// Get takes name of the ingressConfig, and returns the corresponding ingressConfig object, and an error if there is any.
func (c *FakeIngressConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.IngressConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ingressconfigsResource, c.ns, name), &v1alpha1.IngressConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressConfig), err
}

// List takes label and field selectors, and returns the list of IngressConfigs that match those selectors.
func (c *FakeIngressConfigs) List(opts v1.ListOptions) (result *v1alpha1.IngressConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ingressconfigsResource, ingressconfigsKind, c.ns, opts), &v1alpha1.IngressConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.IngressConfigList{}
	for _, item := range obj.(*v1alpha1.IngressConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ingressConfigs.
func (c *FakeIngressConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ingressconfigsResource, c.ns, opts))

}

// Create takes the representation of a ingressConfig and creates it.  Returns the server's representation of the ingressConfig, and an error, if there is any.
func (c *FakeIngressConfigs) Create(ingressConfig *v1alpha1.IngressConfig) (result *v1alpha1.IngressConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ingressconfigsResource, c.ns, ingressConfig), &v1alpha1.IngressConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressConfig), err
}

// Update takes the representation of a ingressConfig and updates it. Returns the server's representation of the ingressConfig, and an error, if there is any.
func (c *FakeIngressConfigs) Update(ingressConfig *v1alpha1.IngressConfig) (result *v1alpha1.IngressConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ingressconfigsResource, c.ns, ingressConfig), &v1alpha1.IngressConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressConfig), err
}

// Delete takes name of the ingressConfig and deletes it. Returns an error if one occurs.
func (c *FakeIngressConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(ingressconfigsResource, c.ns, name), &v1alpha1.IngressConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIngressConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ingressconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.IngressConfigList{})
	return err
}

// Patch applies the patch and returns the patched ingressConfig.
func (c *FakeIngressConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ingressconfigsResource, c.ns, name, data, subresources...), &v1alpha1.IngressConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressConfig), err
}
