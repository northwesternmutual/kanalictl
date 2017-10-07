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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
)

// ValidateAPIKeyBinding performs validation on an APIKeyBinding
func ValidateAPIKeyBinding(binding spec.APIKeyBinding, client utils.HTTPClient, host string) error {
	if binding.Spec.APIProxyName == "" {
		return errors.New("proxy name must be defined")
	}

	// check to make sure that there are no other bindings
	// with the same proxy name
	if err := checkUniqueProxyName(binding.ObjectMeta.Name, binding.ObjectMeta.Namespace, binding.Spec.APIProxyName, client, host); err != nil {
		return err
	}

	return validateKeys(binding.Spec.Keys)
}

func validateKeys(keys []spec.Key) error {

	if keys == nil || len(keys) < 1 {
		return errors.New("must give at least one key permission")
	}

	for _, key := range keys {

		// name must be defined
		if len(key.Name) < 1 {
			return errors.New("key must have a name defined")
		}

		if key.Quota < 0 {
			return errors.New("quota must be non negative")
		}

		// if rate is defined...
		if key.Rate != nil {
			if key.Rate.Amount < 1 {
				return errors.New("rate amount must be a counting number")
			}
			if strings.ToUpper(key.Rate.Unit) != "SECOND" && strings.ToUpper(key.Rate.Unit) != "MINUTE" && strings.ToUpper(key.Rate.Unit) != "HOUR" {
				return errors.New("valid units are 'second', 'minute', 'hour'")
			}
		}

		// validate default rule
		if err := validateRule(key.DefaultRule); err != nil {
			return err
		}

		if err := validateSubpaths(key.Subpaths); err != nil {
			return err
		}

	}

	return nil

}

func validateSubpaths(subpaths []*spec.Path) error {

	for _, subpath := range subpaths {

		if err := checkIfPathIsValid(subpath.Path); err != nil {
			return err
		}

		if err := validateRule(subpath.Rule); err != nil {
			return err
		}

	}

	return nil

}

func validateRule(rule spec.Rule) error {

	if rule.Global && (rule.Granular != nil && rule.Granular.Verbs != nil && len(rule.Granular.Verbs) > 0) {
		return errors.New("global permission granted! granular rules redundant")
	}

	if rule.Granular != nil && rule.Granular.Verbs != nil && len(rule.Granular.Verbs) > 0 {
		for _, verb := range rule.Granular.Verbs {
			switch strings.ToUpper(verb) {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "COPY", "HEAD", "OPTIONS", "LINK", "UNLINK", "PURGE", "LOCK", "UNLOCK", "PROPFIND", "VIEW":
			default:
				return fmt.Errorf("%s is not a valid HTTP verb", strings.ToUpper(verb))
			}
		}
	}

	return nil

}

func checkUniqueProxyName(name, namespace, proxy string, client utils.HTTPClient, host string) error {

	bindingList, err := retrieveBindingList(client, host)
	if err != nil {
		return err
	}

	for _, binding := range bindingList.Bindings {
		if binding.Spec.APIProxyName == proxy {
			if binding.ObjectMeta.Name != name || binding.ObjectMeta.Namespace != namespace {
				return fmt.Errorf("The ApiKeyBinding %s in namespace %s has the same path. Paths must be unique", binding.ObjectMeta.Name, binding.ObjectMeta.Namespace)
			}
		}
	}

	return nil

}

func retrieveBindingList(client utils.HTTPClient, masterHost string) (*spec.APIKeyBindingList, error) {

	resp, err := client.Get(fmt.Sprintf("%s/apis/%s/%s/apikeybindings",
		masterHost,
		APIName,
		APIVersion,
	))
	if err != nil {
		return nil, err
	}

	body := json.NewDecoder(resp.Body)
	list := &spec.APIKeyBindingList{}
	err = body.Decode(list)
	if err != nil {
		return nil, err
	}

	return list, nil

}
