package config

import (
	"github.com/northwesternmutual/kanali/config"
)

var (
	// FlagRSAPublicKeyFile specifies path to RSA public key.
	FlagRSAPublicKeyFile = config.Flag{
		Long:  "rsa.public_key_file",
		Short: "",
		Value: "",
		Usage: "Path to RSA public key.",
	}
)
