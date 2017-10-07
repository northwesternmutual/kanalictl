package cmd

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/northwesternmutual/kanalictl/config"
	"github.com/northwesternmutual/kanalictl/pkg/decrypt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	decryptCmd.Flags().StringP(config.FlagRSAPrivateKeyFile.Long, config.FlagRSAPrivateKeyFile.Short, config.FlagRSAPrivateKeyFile.Value.(string), config.FlagRSAPrivateKeyFile.Usage)
	decryptCmd.Flags().StringP(config.FlagKeyInFile.Long, config.FlagKeyInFile.Short, config.FlagKeyInFile.Value.(string), config.FlagKeyInFile.Usage)

	if err := viper.BindPFlag(config.FlagRSAPrivateKeyFile.Long, decryptCmd.Flags().Lookup(config.FlagRSAPrivateKeyFile.Long)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(config.FlagKeyInFile.Long, decryptCmd.Flags().Lookup(config.FlagKeyInFile.Long)); err != nil {
		panic(err)
	}

	viper.SetDefault(config.FlagRSAPrivateKeyFile.Long, config.FlagRSAPublicKeyFile.Value)
	viper.SetDefault(config.FlagKeyInFile.Long, config.FlagKeyInFile.Value)

	apiKeyCmd.AddCommand(decryptCmd)
}

var decryptCmd = &cobra.Command{
	Use:   `decrypt`,
	Short: `Decrypts API key resources`,
	Long:  `Decrypts API key resources`,
	Run: func(cmd *cobra.Command, args []string) {

		privateKey, err := getPrivateKey(viper.GetString(config.FlagRSAPrivateKeyFile.GetLong()))
		if err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		if err := decrypt.Do(viper.GetString(config.FlagKeyInFile.GetLong()), privateKey); err != nil {
			logrus.Fatalf("%s", err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	},
}

func getPrivateKey(location string) (*rsa.PrivateKey, error) {
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
