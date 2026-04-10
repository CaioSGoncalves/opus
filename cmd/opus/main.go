package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var appName string

func newRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "opus",
		Short: "Simple homelab deployment tool",
	}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy an application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting deploy for:", appName)
			mustRun("go", "test", "./...")
			mustRun("go", "build", "-o", "./bin/"+appName)
			deploySystemdService(appName)
		},
	}

	deployCmd.Flags().StringVar(&appName, "name", "", "Application name")
	deployCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(deployCmd)

	return rootCmd
}

func mustRun(name string, args ...string) {
	fmt.Printf("Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func deploySystemdService(appName string) {
	generateServiceFile(appName)
}
func generateServiceFile(appName string) string {
	return fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
ExecStart=/opt/homelab/%s
WorkingDirectory=/opt/homelab/%s
Restart=always
RestartSec=5

# Limits
MemoryMax=100M
CPUQuota=20%%

[Install]
WantedBy=multi-user.target
`, appName, appName, appName)

}

func main() {
	var rootCmd = newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
