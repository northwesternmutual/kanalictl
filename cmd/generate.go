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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/northwesternmutual/kanalictl/config"
	"github.com/northwesternmutual/kanalictl/pkg/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

func init() {
	generateCmd.Flags().StringP(config.FlagRSAPublicKeyFile.Long, config.FlagRSAPublicKeyFile.Short, config.FlagRSAPublicKeyFile.Value.(string), config.FlagRSAPublicKeyFile.Usage)
	generateCmd.Flags().StringP(config.FlagKeyName.Long, config.FlagKeyName.Short, config.FlagKeyName.Value.(string), config.FlagKeyName.Usage)
	generateCmd.Flags().StringP(config.FlagKeyOutFile.Long, config.FlagKeyOutFile.Short, config.FlagKeyOutFile.Value.(string), config.FlagKeyOutFile.Usage)
	generateCmd.Flags().IntP(config.FlagKeyLength.Long, config.FlagKeyLength.Short, config.FlagKeyLength.Value.(int), config.FlagKeyLength.Usage)
	generateCmd.Flags().StringP(config.FlagKeyData.Long, config.FlagKeyData.Short, config.FlagKeyData.Value.(string), config.FlagKeyData.Usage)

	if err := viper.BindPFlag(config.FlagKeyData.Long, generateCmd.Flags().Lookup(config.FlagKeyData.Long)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(config.FlagRSAPublicKeyFile.Long, generateCmd.Flags().Lookup(config.FlagRSAPublicKeyFile.Long)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(config.FlagKeyName.Long, generateCmd.Flags().Lookup(config.FlagKeyName.Long)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(config.FlagKeyOutFile.Long, generateCmd.Flags().Lookup(config.FlagKeyOutFile.Long)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(config.FlagKeyLength.Long, generateCmd.Flags().Lookup(config.FlagKeyLength.Long)); err != nil {
		panic(err)
	}

	viper.SetDefault(config.FlagKeyData.Long, config.FlagKeyData.Value)
	viper.SetDefault(config.FlagRSAPublicKeyFile.Long, config.FlagRSAPublicKeyFile.Value)
	viper.SetDefault(config.FlagKeyName.Long, config.FlagKeyName.Value)
	viper.SetDefault(config.FlagKeyOutFile.Long, config.FlagKeyOutFile.Value)
	viper.SetDefault(config.FlagKeyLength.Long, config.FlagKeyLength.Value)

	apiKeyCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   `generate`,
	Short: `Creates an API key`,
	Long:  `Creates an API key`,
	Run: func(cmd *cobra.Command, args []string) {

		outFileName, err := getOutFile(viper.GetString(config.FlagKeyOutFile.GetLong()))
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		publicKey, err := getPublicKey(viper.GetString(config.FlagRSAPublicKeyFile.GetLong()))
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		unencryptedKeyData, encryptedKeyData, err := generate.Key(viper.GetString(config.FlagKeyName.GetLong()), viper.GetString(config.FlagKeyData.GetLong()), viper.GetInt(config.FlagKeyLength.GetLong()), publicKey)
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		keyCRD := spec.APIKey{
			TypeMeta: unversioned.TypeMeta{
				APIVersion: "kanali.io/v1",
				Kind:       "ApiKey",
			},
			ObjectMeta: api.ObjectMeta{
				Name:      viper.GetString(config.FlagKeyName.GetLong()),
				Namespace: "default",
			},
			Spec: spec.APIKeySpec{
				APIKeyData: fmt.Sprintf("%x", encryptedKeyData),
			},
		}

		if err := generate.Display(len(outFileName) < 1, unencryptedKeyData, keyCRD); err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		if err := generate.Write(outFileName, keyCRD); err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	},
}

func getOutFile(f string) (string, error) {
	if len(f) < 1 {
		return "", nil
	}

	ext := filepath.Ext(f)
	if len(ext) < 1 {
		return f + ".yaml", nil
	}

	switch ext {
	case ".yaml", ".yml", ".json":
	default:
		return "", errors.New("out file just be either json or yaml format")
	}

	return f, nil
}

func getPublicKey(location string) (pubKey *rsa.PublicKey, returnError error) {

	defer func() {
		if r := recover(); r != nil {
			returnError = errors.New("error parsing public key")
		}
	}()

	var keyBytes []byte

	// read in public key
	keyBytes, err := ioutil.ReadFile(location)
	if err != nil {
		if location == "" {
			return nil, errors.New("public key not specified")
		}
		keyBytes = []byte(location)
	}

	// create a pem block from the public key provided
	block, _ := pem.Decode(keyBytes)
	// parse the pem block into a public key
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	// type assertion
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("error parsing public key")
	}

	return publicKey, nil

}
