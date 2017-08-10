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

func TestValidateAPIKeyBinding(t *testing.T) {

	assert := assert.New(t)
	testBinding := getTestAPIKeyBinding()
	mockClient := &utils.MockClient{}

	assert.Nil(ValidateAPIKeyBinding(testBinding, mockClient, ""), "binding should be valid")

	// test proxy name stuff
	testBinding.Spec.APIProxyName = ""
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "proxy name must be defined")
	testBinding.Spec.APIProxyName = "example-eight"
	testBinding.ObjectMeta.Name = "example-nine"
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "The ApiKeyBinding example-eight in namespace application has the same path. Paths must be unique")
	testBinding.ObjectMeta.Name = "example-eight"

	// test key stuff
	testBinding.Spec.Keys[0].Name = ""
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "key must have a name defined")
	testBinding.Spec.Keys[0].Name = "franks-api-key"
	oldKeys := testBinding.Spec.Keys
	testBinding.Spec.Keys = []spec.Key{}
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "must give at least one key permission")
	testBinding.Spec.Keys = nil
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "must give at least one key permission")
	testBinding.Spec.Keys = oldKeys
	testBinding.Spec.Keys[0].Quota = -1
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "quota must be non negative")
	testBinding.Spec.Keys[0].Quota = 0
	testBinding.Spec.Keys[0].Rate = &spec.Rate{
		Amount: -1,
		Unit:   "minute",
	}
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "rate amount must be a counting number")
	testBinding.Spec.Keys[0].Rate = &spec.Rate{
		Amount: 1,
		Unit:   "frank",
	}
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "valid units are 'second', 'minute', 'hour'")
	testBinding.Spec.Keys[0].Rate = &spec.Rate{
		Amount: 1,
		Unit:   "second",
	}
	testBinding.Spec.Keys[0].Subpaths = []*spec.Path{
		{
			Path: "",
		},
	}
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "path must be defined")
	testBinding.Spec.Keys[0].Subpaths[0].Path = "/"
	assert.Nil(ValidateAPIKeyBinding(testBinding, mockClient, ""), "binding should be valid")
	testBinding.Spec.Keys[0].Subpaths[0].Rule = spec.Rule{
		Global: true,
		Granular: &spec.GranularProxy{
			Verbs: []string{
				"GET",
			},
		},
	}
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "global permission granted! granular rules redundant")
	testBinding.Spec.Keys[0].Subpaths[0].Rule.Global = false
	testBinding.Spec.Keys[0].Subpaths[0].Rule.Granular.Verbs[0] = "frank"
	assert.Equal(ValidateAPIKeyBinding(testBinding, mockClient, "").Error(), "FRANK is not a valid HTTP verb")

}

func getTestAPIKeyBinding() spec.APIKeyBinding {

	return spec.APIKeyBinding{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      "example-eight",
			Namespace: "application",
		},
		Spec: spec.APIKeyBindingSpec{
			APIProxyName: "example-eight",
			Keys: []spec.Key{
				{
					Name: "franks-api-key",
					DefaultRule: spec.Rule{
						Global: true,
					},
				},
			},
		},
	}

}
