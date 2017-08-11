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

	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
)

const (
	// APIName is the Kubernetes TPR resource name
	APIName = "kanali.io"

	// APIVersion is the Kubernetes TPR version
	APIVersion = "v1"
)

// ValidateAPIProxy performs validation on an APIProxy
func ValidateAPIProxy(proxy spec.APIProxy, client utils.HTTPClient, host string) error {

	// is path defined
	if err := checkIfPathIsValid(proxy.Spec.Path); err != nil {
		return err
	}

	// is path unique
	if err := checkUniquePath(proxy.ObjectMeta.Name, proxy.ObjectMeta.Namespace, proxy.Spec.Path, client, host); err != nil {
		return err
	}

	// is target defined
	if err := checkIfTargetIsValid(proxy.Spec.Target); err != nil {
		return err
	}

	// validate hosts
	if err := validateHosts(proxy.Spec.Hosts); err != nil {
		return err
	}

	// validate service
	if err := validateService(proxy.Spec.Service); err != nil {
		return err
	}

	// validate plugins
	if err := validatePlugins(proxy.Spec.Plugins); err != nil {
		return err
	}

	return nil

}

func validateHosts(hosts []spec.Host) error {

	for _, host := range hosts {
		if len(host.SSL.SecretName) < 1 {
			return errors.New("ssl name must be defined")
		}
		if len(host.Name) < 1 {
			return errors.New("host name must be defined if ssl defined")
		}
	}

	return nil

}

func validatePlugins(plugins []spec.Plugin) error {

	for _, plugin := range plugins {
		if len(plugin.Version) < 1 {
			continue
		}
		if len(plugin.Name) < 1 {
			return errors.New("plugin name must be defined if version defined")
		}
	}

	return nil

}

func checkUniquePath(name, namespace, path string, client utils.HTTPClient, host string) error {

	proxyList, err := retrieveProxyList(client, host)
	if err != nil {
		return err
	}

	for _, proxy := range proxyList.Proxies {
		if proxy.Spec.Path == path {
			if proxy.ObjectMeta.Name != name || proxy.ObjectMeta.Namespace != namespace {
				return fmt.Errorf("The ApiProxy %s in namespace %s has the same path. Paths must be unique", proxy.ObjectMeta.Name, proxy.ObjectMeta.Namespace)
			}
		}
	}
	return nil

}

func checkIfPathIsValid(path string) error {
	if path == "" {
		return errors.New("path must be defined")
	}
	if path[0] != '/' {
		return errors.New("path must begin with '/'")
	}
	return nil
}

func checkIfTargetIsValid(path string) error {
	if path == "" {
		return nil
	}
	if path[0] != '/' {
		return errors.New("target must begin with '/'")
	}
	return nil
}

func retrieveProxyList(client utils.HTTPClient, masterHost string) (*spec.APIProxyList, error) {

	resp, err := client.Get(fmt.Sprintf("%s/apis/%s/%s/apiproxies",
		masterHost,
		APIName,
		APIVersion,
	))
	if err != nil {
		return nil, err
	}

	body := json.NewDecoder(resp.Body)
	list := &spec.APIProxyList{}
	err = body.Decode(list)
	if err != nil {
		return nil, err
	}

	return list, nil

}

func validateService(svc spec.Service) error {

	if svc.Port < 1 || svc.Port > 65535 {
		return errors.New("service port must be in range [1-65535]")
	}

	if svc.Name == "" && (svc.Labels == nil || len(svc.Labels) < 1) {
		return errors.New("labels must be defined for dynamic service discovery")
	}

	if svc.Name != "" && svc.Labels != nil && len(svc.Labels) > 0 {
		return errors.New("service name defined, labels are redundant")
	}

	if len(svc.Labels) > 0 {
		return validateLabels(svc.Labels)
	}

	return nil

}

func validateLabels(labels spec.Labels) error {

	for _, label := range labels {
		if label.Name == "" {
			return errors.New("label name must be defined")
		}
		if label.Value == "" && label.Header == "" {
			return errors.New("label must have either a value or a header defined")
		}
		if label.Value != "" && label.Header != "" {
			return errors.New("cannot specify both header and name")
		}
	}
	return nil

}
