/*
Copyright 2020 Giant Swarm GmbH.

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

package v1alpha1

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"

	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
)

// ReleaseCyclesGetter has a method to return a ReleaseCycleInterface.
// A group's client should implement this interface.
type ReleaseCyclesGetter interface {
	ReleaseCycles() ReleaseCycleInterface
}

// ReleaseCycleInterface has methods to work with ReleaseCycle resources.
type ReleaseCycleInterface interface {
	Create(*v1alpha1.ReleaseCycle) (*v1alpha1.ReleaseCycle, error)
	Update(*v1alpha1.ReleaseCycle) (*v1alpha1.ReleaseCycle, error)
	UpdateStatus(*v1alpha1.ReleaseCycle) (*v1alpha1.ReleaseCycle, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ReleaseCycle, error)
	List(opts v1.ListOptions) (*v1alpha1.ReleaseCycleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReleaseCycle, err error)
	ReleaseCycleExpansion
}

// releaseCycles implements ReleaseCycleInterface
type releaseCycles struct {
	client rest.Interface
}

// newReleaseCycles returns a ReleaseCycles
func newReleaseCycles(c *ReleaseV1alpha1Client) *releaseCycles {
	return &releaseCycles{
		client: c.RESTClient(),
	}
}

// Get takes name of the releaseCycle, and returns the corresponding releaseCycle object, and an error if there is any.
func (c *releaseCycles) Get(name string, options v1.GetOptions) (result *v1alpha1.ReleaseCycle, err error) {
	result = &v1alpha1.ReleaseCycle{}
	err = c.client.Get().
		Resource("releasecycles").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ReleaseCycles that match those selectors.
func (c *releaseCycles) List(opts v1.ListOptions) (result *v1alpha1.ReleaseCycleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ReleaseCycleList{}
	err = c.client.Get().
		Resource("releasecycles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested releaseCycles.
func (c *releaseCycles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("releasecycles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a releaseCycle and creates it.  Returns the server's representation of the releaseCycle, and an error, if there is any.
func (c *releaseCycles) Create(releaseCycle *v1alpha1.ReleaseCycle) (result *v1alpha1.ReleaseCycle, err error) {
	result = &v1alpha1.ReleaseCycle{}
	err = c.client.Post().
		Resource("releasecycles").
		Body(releaseCycle).
		Do().
		Into(result)
	return
}

// Update takes the representation of a releaseCycle and updates it. Returns the server's representation of the releaseCycle, and an error, if there is any.
func (c *releaseCycles) Update(releaseCycle *v1alpha1.ReleaseCycle) (result *v1alpha1.ReleaseCycle, err error) {
	result = &v1alpha1.ReleaseCycle{}
	err = c.client.Put().
		Resource("releasecycles").
		Name(releaseCycle.Name).
		Body(releaseCycle).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *releaseCycles) UpdateStatus(releaseCycle *v1alpha1.ReleaseCycle) (result *v1alpha1.ReleaseCycle, err error) {
	result = &v1alpha1.ReleaseCycle{}
	err = c.client.Put().
		Resource("releasecycles").
		Name(releaseCycle.Name).
		SubResource("status").
		Body(releaseCycle).
		Do().
		Into(result)
	return
}

// Delete takes name of the releaseCycle and deletes it. Returns an error if one occurs.
func (c *releaseCycles) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("releasecycles").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *releaseCycles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("releasecycles").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched releaseCycle.
func (c *releaseCycles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReleaseCycle, err error) {
	result = &v1alpha1.ReleaseCycle{}
	err = c.client.Patch(pt).
		Resource("releasecycles").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
