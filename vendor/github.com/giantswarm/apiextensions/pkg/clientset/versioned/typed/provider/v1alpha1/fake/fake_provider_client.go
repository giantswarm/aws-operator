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

package fake

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/clientset/versioned/typed/provider/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeProviderV1alpha1 struct {
	*testing.Fake
}

func (c *FakeProviderV1alpha1) AWSConfigs(namespace string) v1alpha1.AWSConfigInterface {
	return &FakeAWSConfigs{c, namespace}
}

func (c *FakeProviderV1alpha1) AzureConfigs(namespace string) v1alpha1.AzureConfigInterface {
	return &FakeAzureConfigs{c, namespace}
}

func (c *FakeProviderV1alpha1) KVMConfigs(namespace string) v1alpha1.KVMConfigInterface {
	return &FakeKVMConfigs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeProviderV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
