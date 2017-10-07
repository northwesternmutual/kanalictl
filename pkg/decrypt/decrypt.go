package decrypt

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/olekukonko/tablewriter"
	yamlReader "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	keyDataRegex = "^[0-9a-zA-Z]+$"
	label        = "kanali"
)

type result struct {
	name string
	data string
}

// Do attempts to decrypt all API key resources recursively found
// under the specified file or directory.
func Do(inFilePath string, key *rsa.PrivateKey) error {
	fileList, err := discoverFiles(inFilePath)
	if err != nil {
		return err
	}

	renderResults(decryptFiles(fileList, key))

	return nil
}

func renderResults(results []result) {
	if len(results) < 1 {
		fmt.Println("no API keys found")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"API Key Name", "Result"})

	for _, r := range results {
		table.Append([]string{r.name, r.data})
	}

	table.Render()
}

func decryptFiles(fileList []string, key *rsa.PrivateKey) []result {
	allResults := []result{}
	resultsChan := make(chan []result)

	wg := sync.WaitGroup{}
	wg.Add(len(fileList))

	for _, file := range fileList {
		go func(file string) {
			resultsChan <- decryptFile(file, key)
		}(file)
	}
	go func() {
		for r := range resultsChan {
			allResults = append(allResults, r...)
			wg.Done()
		}
	}()

	wg.Wait()
	return allResults
}

func decryptFile(file string, key *rsa.PrivateKey) []result {
	fileResults := []result{}
	fileResultsChan := make(chan result)
	wg := sync.WaitGroup{}

	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return fileResults
	}

	reader := yamlReader.NewYAMLReader(bufio.NewReader(bytes.NewReader(fileData)))

	for {
		doc, err := reader.Read()
		if err != nil && err != io.EOF {
			return fileResults
		} else if err != nil && err == io.EOF {
			break
		}

		wg.Add(1)
		go func(doc []byte) {
			name, data, err := decryptKey(doc, key)
			if err != nil {
				wg.Done()
			} else {
				fileResultsChan <- result{
					name: name,
					data: data,
				}
			}
		}(doc)
		go func() {
			for r := range fileResultsChan {
				fileResults = append(fileResults, r)
				wg.Done()
			}
		}()
	}

	wg.Wait()
	return fileResults
}

func decryptKey(data []byte, key *rsa.PrivateKey) (string, string, error) {
	var meta unversioned.TypeMeta

	err := yaml.Unmarshal(data, &meta)
	if err != nil {
		return "", "", errors.New("yaml document did not contain Kubernetes resource")
	}

	if meta.Kind != "ApiKey" {
		return "", "", errors.New("yaml document was not an ApiKey")
	}

	var apikey spec.APIKey
	if err := yaml.Unmarshal(data, &apikey); err != nil {
		return "", "", err
	}

	cipherText, err := hex.DecodeString(apikey.Spec.APIKeyData)
	if err != nil {
		return "", "", err
	}

	unecryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, cipherText, []byte(label))
	if err != nil {
		return apikey.ObjectMeta.Name, err.Error(), nil
	}
	return apikey.ObjectMeta.Name, string(unecryptedData), nil
}

func discoverFiles(inFilePath string) ([]string, error) {
	fileList := []string{}

	err := filepath.Walk(inFilePath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml", ".json":
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}
