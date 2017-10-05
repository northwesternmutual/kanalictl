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

package utils

import (
  "errors"
	"strings"

  "github.com/ghodss/yaml"
	"github.com/spf13/pflag"
  "k8s.io/kubernetes/pkg/api/unversioned"
)

// GetNamespace retrieves a specific namespace specificed
// in the cli flag set
func GetNamespace(flags *pflag.FlagSet) (string, error) {

	isAll, err := flags.GetBool("all-namespaces")
	namespace, err := flags.GetString("namespace")

	if err != nil {
		return "", err
	}

	if isAll {
		return "", nil
	}

	return namespace, nil

}

// JoinArray joins an array with ','
func JoinArray(arr []string) string {

	if len(arr) < 1 {
		return "none"
	}

	return strings.Join(arr, ",")

}

func GetKind(data []byte) (string, error) {

	var meta unversioned.TypeMeta

	err := yaml.Unmarshal(data, &meta)
	if err != nil {
		return "", errors.New("found a non-kubernetes yaml file")
	}

	return meta.Kind, nil

}
