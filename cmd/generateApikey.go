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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	label       = "kanali"
	apiKeyBytes = 16 // 128 bits
)

func init() {
	generateCmd.AddCommand(generateAPIKeyCmd)
}

var generateAPIKeyCmd = &cobra.Command{
	Use:   `apikey`,
	Short: `generate an api key`,
	Long:  `generate an api key along with the corresponding Kubernetes config containing the encrypted api key`,
	Run: func(cmd *cobra.Command, args []string) {

		outFileName, err := getOutFile(viper.GetString("out-file"))
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		if strings.Compare(viper.GetString("name"), "") == 0 {
			logrus.Fatal("please specify a unique api key name")
			os.Exit(1)
		}

		publicKey, err := getPublicKey(viper.GetString("encrypt_key"))
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		apiKey := createRandByteSlice(apiKeyBytes)

		if len(viper.GetString("apikey")) == 32 {
			apiKey = []byte(viper.GetString("apikey"))
		} else if len(viper.GetString("apikey")) > 0 {
			logrus.Fatalf("apikey must be %d characters long", 32)
			os.Exit(1)
		}

		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, apiKey, []byte(label))
		if err != nil {
			logrus.Fatal("error encrypting your api key - please try again")
			os.Exit(1)
		}

		if _, err := os.Stat(outFileName); !os.IsNotExist(err) {
			// existing file, handle it
			shouldOverwrite(outFileName)
		}

		// print out api key
		if err := displayAPIKey(apiKey); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// apikey struct
		apiKeyStruct := spec.APIKey{
			TypeMeta: unversioned.TypeMeta{
				APIVersion: "kanali.io/v1",
				Kind:       "ApiKey",
			},
			ObjectMeta: api.ObjectMeta{
				Name:      viper.GetString("name"),
				Namespace: viper.GetString("namespace"),
			},
			Spec: spec.APIKeySpec{
				APIKeyData: fmt.Sprintf("%x", ciphertext),
			},
		}

		// write to file if needed
		if err := createFile(outFileName, apiKeyStruct); err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		fmt.Print("Corresponding Kubernetes config written to ")
		c := color.New(color.FgCyan).Add(color.Bold)
		if _, err := c.Println(outFileName); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	},
}

func shouldOverwrite(outFileName string) {

	var input string
	for {
		fmt.Printf("%s exists - do you want to override it? (Y/n) ", outFileName)
		fmt.Scanln(&input)
		if strings.Compare(input, "Y") == 0 || strings.Compare(input, "n") == 0 {
			if strings.Compare(input, "n") == 0 {
				os.Exit(0)
			}
			break
		}
	}

}

func displayAPIKey(key []byte) error {

	fmt.Print("Here is your api key (you will only see this once): ")
	c := color.New(color.FgCyan).Add(color.Bold)

	if len(viper.GetString("apikey")) == 32 {
		if _, err := c.Println(viper.GetString("apikey")); err != nil {
			return err
		}
		return nil
	}

	if _, err := c.Println(hex.EncodeToString(key)); err != nil {
		return err
	}

	return nil

}

func createRandByteSlice(length int) []byte {

	s := make([]byte, length)
	_, err := rand.Read(s)
	if err != nil {
		logrus.Fatal("error reading from entropy source to generate api key")
		os.Exit(1)
	}

	return s

}

func createFile(f string, key spec.APIKey) error {

	// handle yaml
	if strings.Compare(strings.Split(f, ".")[1], "yaml") == 0 || strings.Compare(strings.Split(f, ".")[1], "yml") == 0 {
		y, err := yaml.Marshal(key)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(f, y, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// handle json
	j, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f, j, 0644)
	if err != nil {
		return err
	}
	return nil

}

func getOutFile(f string) (string, error) {

	arr := strings.Split(f, ".")

	switch len(arr) {
	case 1:
		return arr[0] + ".yaml", nil
	case 2:
		if !(strings.Compare(arr[1], "json") == 0 || strings.Compare(arr[1], "yaml") == 0 || strings.Compare(arr[1], "yml") == 0) {
			return "", errors.New("out file just be in json or yaml format")
		}
		return f, nil
	default:
		return "", errors.New("please double check the format of your outfile and try again")

	}

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
