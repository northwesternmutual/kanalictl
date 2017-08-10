// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package validation

import (
	"testing"

	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

func TestValidateAPIProxy(t *testing.T) {

	assert := assert.New(t)
	testProxy := getTestAPIProxy()
	mockClient := &utils.MockClient{}

	assert.Nil(ValidateAPIProxy(testProxy, mockClient, ""), "proxy should be valid")

	// test path stuff
	testProxy.Spec.Path = ""
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "path must be defined")
	testProxy.Spec.Path = "api"
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "path must begin with '/'")
	testProxy.Spec.Path = "/api/v1/example-one"
	testProxy.ObjectMeta.Namespace = "namespace-two"
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "The ApiProxy example-one in namespace application has the same path. Paths must be unique")
	testProxy.ObjectMeta.Namespace = "application"

	// test target stuff
	testProxy.Spec.Target = "/"
	assert.Nil(ValidateAPIProxy(testProxy, mockClient, ""), "proxy should be valid")
	testProxy.Spec.Target = "api"
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "target must begin with '/'")
	testProxy.Spec.Target = ""

	// test host stuff
	testProxy.Spec.Hosts[1] = spec.Host{SSL: spec.SSL{SecretName: "mySecretOne"}}
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "host name must be defined if ssl defined")
	testProxy.Spec.Hosts[1] = spec.Host{Name: "bar.foo.com"}

	// test service stuff
	testProxy.Spec.Service.Port = -1
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "service port must be in range [1-65535]")
	testProxy.Spec.Service.Port = 65536
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "service port must be in range [1-65535]")
	testProxy.Spec.Service.Port = 8080
	testProxy.Spec.Service.Name = ""
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "labels must be defined for dynamic service discovery")
	testProxy.Spec.Service.Name = "my-service"
	testProxy.Spec.Service.Labels = []spec.Label{{Name: "label-one", Value: "value-one"}}
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "service name defined, labels are redundant")
	testProxy.Spec.Service.Name = ""
	testProxy.Spec.Service.Labels = []spec.Label{{Name: "", Value: "value-one"}}
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "label name must be defined")
	testProxy.Spec.Service.Labels = []spec.Label{{Name: "label-one"}}
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "label must have either a value or a header defined")
	testProxy.Spec.Service.Labels = []spec.Label{{Name: "label-one", Value: "value-one", Header: "header-one"}}
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "cannot specify both header and name")
	testProxy.Spec.Service.Labels = []spec.Label{{Name: "label-one", Value: "value-one"}}
	assert.Nil(ValidateAPIProxy(testProxy, mockClient, ""), "proxy should be valid")

	// validate plugin stuff
	testProxy.Spec.Plugins[0].Name = ""
	assert.Equal(ValidateAPIProxy(testProxy, mockClient, "").Error(), "plugin name must be defined if version defined")

}

func getTestAPIProxy() spec.APIProxy {

	return spec.APIProxy{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      "example-one",
			Namespace: "application",
		},
		Spec: spec.APIProxySpec{
			Path: "/api/v1/example-one",
			Hosts: []spec.Host{
				{
					Name: "foo.bar.com",
					SSL: spec.SSL{
						SecretName: "mySecretTwo",
					},
				},
				{
					Name: "bar.foo.com",
				},
			},
			Service: spec.Service{
				Name: "my-service",
				Port: 8080,
			},
			Plugins: []spec.Plugin{
				{
					Name:    "apikey",
					Version: "1.0.0",
				},
				{
					Name: "jwt",
				},
			},
			SSL: spec.SSL{
				SecretName: "mySecret",
			},
		},
	}

}
