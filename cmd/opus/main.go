package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	LocalTempBaseDir     = "/tmp"
	RemoteTempBaseDir    = "/tmp"
	RemoteBinBaseDir     = "/opt/services"
	RemoteSystemdBaseDir = "/etc/systemd/system"
)

//go:embed svc.tmpl
var SvcTemplateRaw string

var (
	RemoteHost string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "opus",
		Short: "Simple service deployment tool",
	}
	rootCmd.AddCommand(initSetupCmd())
	rootCmd.AddCommand(initDeployCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
