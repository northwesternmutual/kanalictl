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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	generateCmd.PersistentFlags().StringP("name", "n", "", "name of the api key")
	generateCmd.PersistentFlags().StringP("namespace", "s", "default", "the namespace the api key should belong to")
	generateCmd.PersistentFlags().StringP("encrypt_key", "k", "", "location of public key to use during encryption")
	generateCmd.PersistentFlags().StringP("out-file", "o", "", "location to save Kubernetes config to")
	generateCmd.PersistentFlags().StringP("apikey", "a", "", "encrypt existing apikey")

	if err := viper.BindPFlag("name", generateCmd.PersistentFlags().Lookup("name")); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := viper.BindPFlag("namespace", generateCmd.PersistentFlags().Lookup("namespace")); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := viper.BindPFlag("encrypt_key", generateCmd.PersistentFlags().Lookup("encrypt_key")); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := viper.BindPFlag("out-file", generateCmd.PersistentFlags().Lookup("out-file")); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := viper.BindPFlag("apikey", generateCmd.PersistentFlags().Lookup("apikey")); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	viper.SetDefault("namespace", "default")
	viper.SetDefault("encrypt_key", "")

	RootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   `generate`,
	Short: `generate an api key`,
	Long:  `generate an api key along with the corresponding Kubernetes config containing the encrypted api key`,
}
