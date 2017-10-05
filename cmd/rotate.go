package cmd

import (
  "os"
  "fmt"

  "github.com/spf13/cobra"
  "github.com/northwesternmutual/kanalictl/pkg/rotate"
)

func init() {
	rotateCmd.Flags().StringP("pub_key", "n", "", "Location to the new RSA public key file to use.")
  rotateCmd.Flags().StringP("private_key", "k", "", "Location to the current RSA private key file to use.")
  rotateCmd.Flags().StringP("file", "f", "", "Location of the file or directory that contain APIKey spec files.")
	RootCmd.AddCommand(rotateCmd)
}

var rotateCmd = &cobra.Command{
	Use:   `rotate`,
	Short: `Rotate RSA encrypted API keys.`,
	Long:  `Rotate RSA encrypted API keys.`,
	Run: rotateCmdRun,
}

var rotateCmdRun = func(cmd *cobra.Command, args []string) {
  publicKey, err := cmd.Flags().GetString("pub_key")
  privateKey, err := cmd.Flags().GetString("private_key")
  configPath, err := cmd.Flags().GetString("file")
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }

  if err := rotate.Rotate(publicKey, privateKey, configPath); err != nil {
    fmt.Print(err.Error())
    os.Exit(1)
  }
  os.Exit(0)
}