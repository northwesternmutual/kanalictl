package rotate

import (
  "io"
  "os"
  "fmt"
	"bufio"
	"bytes"
  "errors"
  "io/ioutil"
  "crypto/rsa"
  "crypto/rand"
	"crypto/sha256"
  "crypto/x509"
	"path/filepath"
  "encoding/hex"
  "encoding/pem"

	"github.com/ghodss/yaml"
  "github.com/olekukonko/tablewriter"
	"github.com/northwesternmutual/kanali/spec"
  "github.com/northwesternmutual/kanalictl/utils"
	"github.com/northwesternmutual/kanalictl/validation"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
)

func Rotate(publicKeyPath, privateKeyPath, filePath string) error {

	if publicKeyPath == "" {
		return errors.New("path to rsa public key must be specified")
	}
	if privateKeyPath == "" {
		return errors.New("path to rsa private key must be specified")
	}
	if filePath == "" {
		return errors.New("path to APIKey specs must be specified")
	}

  decryptionKey, err := loadDecryptionKey(privateKeyPath)
  if err != nil {
    return err
  }

  publicKey, err := utils.GetPublicKey(publicKeyPath)
  if err != nil {
    return err
  }

  fileList := []string{}
  fileErrorList := []string{}

  err = filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
      fileList = append(fileList, path)
    }
    return nil
  })

  jobsReport := &jobs{}

  for _, file := range fileList {
    jbs, err := preformRotation(file, decryptionKey, publicKey)
    if err != nil {
      fileErrorList = append(fileErrorList, file)
    }
    jobsReport.add((*jbs)...)
  }

  table := tablewriter.NewWriter(os.Stdout)
  table.SetAutoWrapText(false)
  table.SetHeader([]string{"TOTAL", "SUCCESS", "ERROR"})
  numSuccess := 0
  for _, j := range *jobsReport {
    if j.err == nil {
      numSuccess++
    }
	}
  table.Append([]string{fmt.Sprintf("%d", len(*jobsReport)), fmt.Sprintf("%d", numSuccess), fmt.Sprintf("%d", len(*jobsReport)-numSuccess)})
  table.Render()

  table = tablewriter.NewWriter(os.Stdout)
  table.SetAutoWrapText(false)
  table.SetHeader([]string{"NAME", "NAMESPACE", "FILE", "ROTATED", "ERROR"})

	for _, j := range *jobsReport {
    var errString string
    var wasRotated string
    if j.err == nil {
      errString = "none"
      wasRotated = "true"
    } else {
      errString = j.err.Error()
      wasRotated = "false"
    }
    if j.isAPIKey {
      table.Append([]string{j.name, j.namespace, j.file, wasRotated, errString})
    }
	}

  table.Render()

  return nil

}

func preformRotation(path string, decryptionKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (*jobs, error) {
  jbs := &jobs{}
  numAPIKeys := 0

  yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		return jbs, err
	}

  yamlReader := k8sYaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(yamlData)))

  for {
    document, err := yamlReader.Read()
    if err != nil && err != io.EOF {
      return jbs, err
    } else if err != nil && err == io.EOF {
      break
    }
    currJob := attemptAPIKeyRotation(document, decryptionKey, publicKey)
    if currJob.isAPIKey {
      numAPIKeys++
    }
    currJob.file = path
    jbs.add(currJob)
  }

  if len(*jbs) < 1 {
    return jbs, nil
  } else if numAPIKeys < 1 {
    return jbs, nil
  }

  f, err := os.Create(getStagedFilePath(path))
  if err != nil {
    return jbs, err
  }
  defer f.Close()
  w := bufio.NewWriter(f)

  var wasError bool

  for i, item := range *jbs {
    if i != 0 {
      _, err := w.Write([]byte("\n---\n"))
      if err != nil {
        wasError = true
      }
    }
    var num int
    var err error
    if item.newData == nil {
      num, err = w.Write(item.oldData)
    } else {
      num, err = w.Write(item.newData)
    }
    if err != nil {
      wasError = true
      (&item).err = err
    }
    (&item).bytesWritten = num
  }
  w.Flush()

  if !wasError {
    os.Rename(getStagedFilePath(path), path)
  }

  return jbs, nil
}

func attemptAPIKeyRotation(data []byte, decryptionKey *rsa.PrivateKey, publicKey *rsa.PublicKey) job {
  tmpJob := job{
    oldData: data,
  }

  kind, err := utils.GetKind(data)
	if err != nil {
    tmpJob.err = err
    return tmpJob
	}

	if kind != "ApiKey" {
		return tmpJob
	} else {
    tmpJob.isAPIKey = true
  }

	var apikey spec.APIKey
	if err := yaml.Unmarshal(data, &apikey); err != nil {
    tmpJob.err = err
		return tmpJob
	} else {
    tmpJob.name = apikey.ObjectMeta.Name
    tmpJob.namespace = apikey.ObjectMeta.Namespace
  }

	if err := validation.ValidateAPIKey(apikey); err != nil {
    tmpJob.err = err
		return tmpJob
	}

  unecryptedData, err := decrypt(apikey.Spec.APIKeyData, decryptionKey)
  if err != nil {
    tmpJob.err = err
		return tmpJob
  }

  fmt.Printf("%x", unecryptedData)
  fmt.Println(unecryptedData)

  ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, unecryptedData, []byte(utils.LABEL))
  if err != nil {
    tmpJob.err = err
		return tmpJob
  }

  apikey.Spec.APIKeyData = fmt.Sprintf("%x", ciphertext)

  newData, err := yaml.Marshal(apikey)
  if err != nil {
    tmpJob.err = err
		return tmpJob
  } else {
    tmpJob.newData = newData
  }

  return tmpJob

}

func getStagedFilePath(path string) string {
  name := filepath.Base(path)
  name = name + "_staged"
  return filepath.Join(filepath.Dir(path), name)
}

func loadDecryptionKey(location string) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func decrypt(data string, key *rsa.PrivateKey) ([]byte, error) {
	cipherText, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}
	unencryptedAPIKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, cipherText, []byte(utils.LABEL))
	if err != nil {
		return nil, err
	}

	return unencryptedAPIKey, nil
}