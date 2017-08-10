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

func init() {
	getCmd.AddCommand(getAPIKeyBindingsCmd)
}

var getAPIKeyBindingsCmd = &cobra.Command{
	Use:   `apikeybindings`,
	Short: `Get all apikeybindings in one or every namespace.`,
	Long:  `Get all apikeybindings in one or every namespace.`,
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

		bindingList, err := getAPIKeyBindings(ctlr, namespace)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if len(bindingList.Bindings) < 1 {
			fmt.Println("No resources found.")
			os.Exit(0)
		}

		output, err := cmd.Flags().GetString("out")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if err := printBindings(bindingList.Bindings, output, namespace); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	},
}

func getAPIKeyBindings(ctlr *controller.Controller, namespace string) (*spec.APIKeyBindingList, error) {

	url := fmt.Sprintf("apis/%s/%s/apikeybindings", APIName, APIVersion)

	if len(namespace) > 0 {
		url = fmt.Sprintf("apis/%s/%s/namespaces/%s/apikeybindings", APIName, APIVersion, namespace)
	}

	resp, err := ctlr.RestClient.Client.Get(fmt.Sprintf("%s/%s", ctlr.MasterHost, url))
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

func printBindings(bindings []spec.APIKeyBinding, output, namespace string) error {

	switch strings.ToUpper(output) {
	case "":
		return doPrintTextBindings(bindings, namespace)
	case "YAML":
		return doPrintYamlBindings(bindings)
	case "JSON":
		return doPrintJSONBindings(bindings)
	default:
		return errors.New("wrong output format - please use only json|yaml")
	}
}

func doPrintTextBindings(bindings []spec.APIKeyBinding, namespace string) error {

	table := uitable.New()
	table.MaxColWidth = 100

	if namespace == "" {
		table.AddRow("NAME", "PROXY", "API_KEYS")
	} else {
		table.AddRow("NAMESPACE", "NAME", "PROXY", "API_KEYS")
	}

	for _, binding := range bindings {

		apiKeysList := []string{}

		for _, key := range binding.Spec.Keys {
			apiKeysList = append(apiKeysList, key.Name)
		}

		if namespace == "" {
			table.AddRow(binding.ObjectMeta.Name, binding.Spec.APIProxyName, utils.JoinArray(apiKeysList))
		} else {
			table.AddRow(namespace, binding.ObjectMeta.Name, binding.Spec.APIProxyName, utils.JoinArray(apiKeysList))
		}

	}

	fmt.Println(table.String())

	return nil

}

func doPrintYamlBindings(bindings []spec.APIKeyBinding) error {

	for _, binding := range bindings {

		b, err := yaml.Marshal(binding)
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}

func doPrintJSONBindings(bindings []spec.APIKeyBinding) error {

	for _, binding := range bindings {

		b, err := json.MarshalIndent(binding, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}
