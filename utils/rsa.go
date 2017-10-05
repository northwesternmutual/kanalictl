package utils

import (
  "errors"
  "io/ioutil"
  "crypto/x509"
  "crypto/rsa"
  "encoding/pem"
)

const (
  LABEL = "kanali"
)

func GetPublicKey(location string) (pubKey *rsa.PublicKey, returnError error) {

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
