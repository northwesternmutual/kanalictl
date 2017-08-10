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

	"github.com/northwesternmutual/kanali/controller"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/utils"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	getCmd.AddCommand(getAPIKeiesCmd)
}

var getAPIKeiesCmd = &cobra.Command{
	Use:   `apikeies`,
	Short: `Get all apikeies in one or every namespace.`,
	Long:  `Get all apikeies in one or every namespace.`,
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

		keyList, err := getAPIKeys(ctlr, namespace)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if len(keyList.Keys) < 1 {
			fmt.Println("No resources found.")
			os.Exit(0)
		}

		output, err := cmd.Flags().GetString("out")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if err := printKeys(keyList.Keys, output, namespace); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	},
}

func getAPIKeys(ctlr *controller.Controller, namespace string) (*spec.APIKeyList, error) {

	url := fmt.Sprintf("apis/%s/%s/apikeies", APIName, APIVersion)

	if len(namespace) > 0 {
		url = fmt.Sprintf("apis/%s/%s/namespaces/%s/apikeies", APIName, APIVersion, namespace)
	}

	resp, err := ctlr.RestClient.Client.Get(fmt.Sprintf("%s/%s", ctlr.MasterHost, url))
	if err != nil {
		return nil, err
	}

	body := json.NewDecoder(resp.Body)
	list := &spec.APIKeyList{}
	err = body.Decode(list)
	if err != nil {
		return nil, err
	}

	return list, nil

}

func printKeys(keys []spec.APIKey, output, namespace string) error {

	switch strings.ToUpper(output) {
	case "":
		return doPrintTextKeys(keys, namespace)
	case "YAML":
		return doPrintYamlKeys(keys)
	case "JSON":
		return doPrintJSONKeys(keys)
	default:
		return errors.New("wrong output format - please use only json|yaml")
	}
}

func doPrintTextKeys(keys []spec.APIKey, namespace string) error {

	table := uitable.New()
	table.MaxColWidth = 30

	if namespace == "" {
		table.AddRow("NAME", "DATA")
	} else {
		table.AddRow("NAMESPACE", "NAME", "DATA")
	}

	for _, key := range keys {

		if namespace == "" {
			table.AddRow(key.ObjectMeta.Name, key.Spec.APIKeyData)
		} else {
			table.AddRow(namespace, key.ObjectMeta.Name, key.Spec.APIKeyData)
		}

	}

	fmt.Println(table.String())

	return nil

}

func doPrintYamlKeys(keys []spec.APIKey) error {

	for _, key := range keys {

		b, err := yaml.Marshal(key)
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}

func doPrintJSONKeys(keys []spec.APIKey) error {

	for _, key := range keys {

		b, err := json.MarshalIndent(key, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	}

	return nil

}
