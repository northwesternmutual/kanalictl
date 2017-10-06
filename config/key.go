package config

import (
	"github.com/northwesternmutual/kanali/config"
)

var (
	// FlagKeyName specifies the API key name.
	FlagKeyName = config.Flag{
		Long:  "key.name",
		Short: "",
		Value: "",
		Usage: "Name of API data.",
	}
	// FlagKeyData specifies existing API key data.
	FlagKeyData = config.Flag{
		Long:  "key.data",
		Short: "",
		Value: "",
		Usage: "Existing API key data.",
	}
	// FlagKeyOutFile specifies path to RSA public key.
	FlagKeyOutFile = config.Flag{
		Long:  "key.out_file",
		Short: "",
		Value: "",
		Usage: "Existing API key data.",
	}
	// FlagKeyLength specifies the desired length of the generated API key.
	FlagKeyLength = config.Flag{
		Long:  "key.length",
		Short: "",
		Value: 32,
		Usage: "Existing API key data.",
	}
)
