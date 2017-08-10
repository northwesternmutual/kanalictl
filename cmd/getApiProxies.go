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

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/northwesternmutual/kanali/controller"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
	"github.com/spf13/cobra"
)

const (
	// APIName is the Kubernetes TPR resource name
	APIName = "kanali.io"

	// APIVersion is the Kubernetes TPR version
	APIVersion = "v1"
)

func init() {
	getCmd.AddCommand(getAPIProxiesCmd)
}

var getAPIProxiesCmd = &cobra.Command{
	Use:   `apiproxies`,
	Short: `Get all apiproxies in one or every namespace.`,
	Long:  `Get all apiproxies in one or every namespace.`,
	Run: func(cmd *cobra.Command, args []string) {

		ctlr, err := controller.New()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		namespace, err := utils.GetNamespace(cmd.Flags())
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		proxyList, err := getAPIProxies(ctlr, namespace)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if len(proxyList.Proxies) < 1 {
			fmt.Println("No resources found.")
			os.Exit(0)
		}

		output, err := cmd.Flags().GetString("out")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if err := printProxies(proxyList.Proxies, output, namespace); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	},
}

func getAPIProxies(ctlr *controller.Controller, namespace string) (*spec.APIProxyList, error) {

	url := fmt.Sprintf("apis/%s/%s/apiproxies", APIName, APIVersion)

	if len(namespace) > 0 {
		url = fmt.Sprintf("apis/%s/%s/namespaces/%s/apiproxies", APIName, APIVersion, namespace)
	}

	resp, err := ctlr.RestClient.Client.Get(fmt.Sprintf("%s/%s", ctlr.MasterHost, url))
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

func printProxies(proxies []spec.APIProxy, output, namespace string) error {

	switch strings.ToUpper(output) {
	case "":
		return doPrintTextProxies(proxies, namespace)
	case "YAML":
		return doPrintYamlProxies(proxies)
	case "JSON":
		return doPrintJSONProxies(proxies)
	default:
		return errors.New("wrong output format. Please use only json|yaml")
	}
}

func doPrintTextProxies(proxies []spec.APIProxy, namespace string) error {

	table := uitable.New()
	table.MaxColWidth = 100

	if namespace == "" {
		table.AddRow("NAME", "PATH", "TARGET", "HOSTS", "SERVICE", "PLUGINS", "SSL")
	} else {
		table.AddRow("NAMESPACE", "NAME", "PATH", "TARGET", "HOSTS", "SERVICE", "PLUGINS", "SSL")
	}

	for _, proxy := range proxies {

		pluginsList := []string{}
		hostsList := []string{}
		service := ""
		ssl := "disabled"

		for _, plugin := range proxy.Spec.Plugins {
			pluginsList = append(pluginsList, plugin.Name)
		}

		if proxy.Spec.SSL != (spec.SSL{}) {
			ssl = "enabled"
		}

		for _, host := range proxy.Spec.Hosts {
			hostsList = append(hostsList, host.Name)
			if ssl != "disabled" {
				continue
			}
			if host.SSL != (spec.SSL{}) {
				ssl = "conditionally enabled"
			}
		}

		if proxy.Spec.Service.Name == "" {
			service = "dynamic"
		} else {
			service = fmt.Sprintf("%s.%s.svc.cluster.local:%d", proxy.ObjectMeta.Name, proxy.ObjectMeta.Namespace, proxy.Spec.Service.Port)
		}

		if namespace == "" {
			table.AddRow(proxy.ObjectMeta.Name, proxy.Spec.Path, proxy.Spec.Target, utils.JoinArray(hostsList), service, utils.JoinArray(pluginsList), ssl)
		} else {
			table.AddRow(namespace, proxy.ObjectMeta.Name, proxy.Spec.Path, proxy.Spec.Target, utils.JoinArray(hostsList), service, utils.JoinArray(pluginsList), ssl)
		}

	}

	fmt.Println(table.String())

	return nil

}

func doPrintYamlProxies(proxies []spec.APIProxy) error {

	for _, proxy := range proxies {

		b, err := yaml.Marshal(proxy)
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}

func doPrintJSONProxies(proxies []spec.APIProxy) error {

	for _, proxy := range proxies {

		b, err := json.MarshalIndent(proxy, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}
