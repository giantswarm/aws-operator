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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/clientset/versioned/typed/core/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCoreV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCoreV1alpha1) CertConfigs(namespace string) v1alpha1.CertConfigInterface {
	return &FakeCertConfigs{c, namespace}
}

func (c *FakeCoreV1alpha1) DraughtsmanConfigs(namespace string) v1alpha1.DraughtsmanConfigInterface {
	return &FakeDraughtsmanConfigs{c, namespace}
}

func (c *FakeCoreV1alpha1) FlannelConfigs(namespace string) v1alpha1.FlannelConfigInterface {
	return &FakeFlannelConfigs{c, namespace}
}

func (c *FakeCoreV1alpha1) IngressConfigs(namespace string) v1alpha1.IngressConfigInterface {
	return &FakeIngressConfigs{c, namespace}
}

func (c *FakeCoreV1alpha1) NodeConfigs(namespace string) v1alpha1.NodeConfigInterface {
	return &FakeNodeConfigs{c, namespace}
}

func (c *FakeCoreV1alpha1) StorageConfigs(namespace string) v1alpha1.StorageConfigInterface {
	return &FakeStorageConfigs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCoreV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
