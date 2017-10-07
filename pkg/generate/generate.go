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

package generate

import (
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	mathRand "math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/ghodss/yaml"
	"github.com/northwesternmutual/kanali/spec"
)

var (
	src = mathRand.NewSource(time.Now().UnixNano())
)

const (
	keyNameRegex  = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	keyDataRegex  = "^[0-9a-zA-Z]+$"
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	label         = "kanali"
)

// Key generate an encrypted key. It produces both the
// unencrypted key as well as the encrypted key data.
func Key(keyName, existingKey string, length int, encryptKey *rsa.PublicKey) ([]byte, []byte, error) {

	if !regexp.MustCompile(keyNameRegex).MatchString(keyName) {
		return nil, nil, fmt.Errorf("key name must conform to the pattern %s", keyNameRegex)
	}

	unencryptedKeyData, err := generateKeyData(existingKey, length)
	if err != nil {
		return nil, nil, err
	}

	encryptedKeyData, err := encryptKeyData(unencryptedKeyData, encryptKey)
	if err != nil {
		return nil, nil, err
	}

	return unencryptedKeyData, encryptedKeyData, nil

}

// Display will display an API key an the corresponding Kubernetes config
func Display(showCRD bool, unencryptedKeyData []byte, keyCRD spec.APIKey) error {

	fmt.Printf("Here is your api key (you will only see this once): %s\n", string(unencryptedKeyData))

	if !showCRD {
		return nil
	}

	yamlData, err := yaml.Marshal(keyCRD)
	if err != nil {
		return err
	}

	fmt.Printf("Corresponding Kubernetes config:\n%s", string(yamlData))
	return nil
}

// Write will write the corresponding Kubernetes config to a file.
func Write(outFileName string, keyCRD spec.APIKey) error {
	if len(outFileName) < 1 {
		return nil
	}

	_, err := os.Stat(outFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if !os.IsNotExist(err) && shouldNotOverwrite(outFileName) {
		return nil
	}

	var fileData []byte

	if filepath.Ext(outFileName) != ".json" {
		yamlData, err := yaml.Marshal(keyCRD)
		if err != nil {
			return err
		}
		fileData = yamlData
	} else {
		jsonData, err := json.MarshalIndent(keyCRD, "", "   ")
		if err != nil {
			return err
		}
		fileData = jsonData
	}

	if err := ioutil.WriteFile(outFileName, fileData, 0644); err != nil {
		return err
	}

	fmt.Printf("Corresponding Kubernetes config written to %s\n", outFileName)
	return nil
}

func shouldNotOverwrite(outFileName string) bool {
	var input string

	for {
		fmt.Printf("%s exists - do you want to override it? (Y/n) ", outFileName)
		fmt.Scanln(&input)
		switch input {
		case "Y":
			return false
		case "n":
			return true
		}
	}
}

func encryptKeyData(unencryptedKeyData []byte, encryptKey *rsa.PublicKey) ([]byte, error) {
	if encryptKey == nil {
		return nil, errors.New("no public key provided")
	}

	return rsa.EncryptOAEP(sha256.New(), cryptoRand.Reader, encryptKey, unencryptedKeyData, []byte(label))
}

func generateKeyData(existingKey string, length int) ([]byte, error) {
	if len(existingKey) > 0 {
		if !regexp.MustCompile(keyDataRegex).MatchString(existingKey) {
			return nil, fmt.Errorf("key data must conform to the pattern %s", keyDataRegex)
		}

		return []byte(existingKey), nil
	}

	if len(existingKey) < 1 && length < 1 {
		return nil, errors.New("key length must be an greater than zero")
	}

	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b, nil
}
