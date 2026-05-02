package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func initUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use [target]",
		Short: "Switch kubeconfig to target",
		Args:  cobra.ExactArgs(1),
		Run:   runUse,
	}

	return cmd
}

func runUse(cmd *cobra.Command, args []string) {
	target := args[0]

	// allow: opus use default → go back to normal kubeconfig
	if target == "default" {
		fmt.Println("unset KUBECONFIG")
		return
	}

	kubeconfig := filepath.Join(
		os.Getenv("HOME"),
		".kube",
		"config-opus-"+target,
	)

	if _, err := os.Stat(kubeconfig); err != nil {
		fmt.Fprintf(os.Stderr, "kubeconfig not found: %s\n", kubeconfig)
		os.Exit(1)
	}

	// this is evaluated by the shell function
	fmt.Printf("export KUBECONFIG=%s\n", kubeconfig)
	fmt.Printf("echo 'using %s'\n", target)
}
