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

package controller

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/northwesternmutual/kanali/controller"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
	"github.com/northwesternmutual/kanalictl/validation"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// CreateOrApply validates a spec and then performs either a create or apply
func CreateOrApply(op, path string) (string, int) {

	// check if file was passed in
	if path == "" {
		return "file must be specified", 1
	}

	// turn potential relative path into absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err.Error(), 1
	}

	// attempt to parse YAML file
	yamlData, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "file is not a valid YAML file", 1
	}

	kind, err := getKind(yamlData)
	if err != nil {
		return err.Error(), 1
	}

	switch kind {
	case "APIKey":
		if err := handleAPIKey(yamlData); err != nil {
			return err.Error(), 1
		}
	case "APIProxy":
		if err := handleAPIProxy(yamlData); err != nil {
			return err.Error(), 1
		}
	case "APIKeyBinding":
		if err := handleAPIKeyBinding(yamlData); err != nil {
			return err.Error(), 1
		}
	default:
		return "please use kubectl for this configuration file", 1
	}

	return utils.Execute("kubectl", op, "-f", path)

}

func handleAPIKeyBinding(data []byte) error {

	var binding spec.APIKeyBinding

	err := yaml.Unmarshal(data, &binding)
	if err != nil {
		return err
	}

	ctlr, err := controller.New()
	if err != nil {
		return err
	}

	if err := validation.ValidateAPIKeyBinding(binding, ctlr.RestClient.Client, ctlr.MasterHost); err != nil {
		return err
	}

	return nil

}

func handleAPIProxy(data []byte) error {

	var proxy spec.APIProxy

	err := yaml.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}

	ctlr, err := controller.New()
	if err != nil {
		return err
	}

	if err := validation.ValidateAPIProxy(proxy, ctlr.RestClient.Client, ctlr.MasterHost); err != nil {
		return err
	}

	return nil

}

func handleAPIKey(data []byte) error {

	var apikey spec.APIKey

	err := yaml.Unmarshal(data, &apikey)
	if err != nil {
		return err
	}
	if err := validation.ValidateAPIKey(apikey); err != nil {
		return err
	}

	return nil

}

func getKind(data []byte) (string, error) {

	var meta unversioned.TypeMeta

	err := yaml.Unmarshal(data, &meta)
	if err != nil {
		return "", errors.New("file is not a valid Kubernetes configuration file")
	}

	return meta.Kind, nil

}
