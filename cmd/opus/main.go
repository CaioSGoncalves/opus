package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	serviceName   string
	remoteHost    string
	remoteBaseDir = "/opt/services"
)

//go:embed svc.tmpl
var svcTemplateRaw string

func newRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "opus",
		Short: "Simple service deployment tool",
	}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy a Go service",
		Run:   runDeploy,
	}

	deployCmd.Flags().StringVar(&serviceName, "name", "", "Service name")
	deployCmd.Flags().StringVar(&remoteHost, "host", "", "Remote host user@host")
	deployCmd.MarkFlagRequired("name")
	deployCmd.MarkFlagRequired("host")

	rootCmd.AddCommand(deployCmd)

	return rootCmd
}

func runDeploy(cmd *cobra.Command, args []string) {
	fmt.Println("Starting deploy for:", serviceName)

	// build Go binary
	localDir := filepath.Join("/tmp", serviceName)
	binPath := filepath.Join(localDir, serviceName)
	pkgDir := "./" + filepath.Join("cmd", serviceName)
	mustRun("go", "test", "./...")
	mustRun("env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", binPath, pkgDir)

	// deploy to remote server
	remoteTempDir := localDir
	remoteDir := filepath.Join(remoteBaseDir, serviceName)

	generateServiceFile(localDir, serviceName, remoteDir)
	mustRun("scp", "-r", localDir, remoteHost+":"+remoteTempDir)

	sshCmd := fmt.Sprintf(
		"sudo bash -c 'mkdir -p %[1]s && rm -rf %[1]s/* && mv %[2]s/* %[1]s && chmod +x %[1]s && rm -rf %[2]s'",
		remoteDir,
		remoteTempDir,
	)
	mustRun("ssh", "-t", remoteHost, sshCmd)

	err := os.RemoveAll(localDir)
	if err != nil {
		log.Fatal(err)
	}
}

func mustRun(name string, args ...string) {
	fmt.Printf("Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func generateServiceFile(localDir string, svcName string, remoteDir string) {
	err := os.MkdirAll(localDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(filepath.Join(localDir, svcName+".service"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	svcTempl, err := template.New("service").Parse(svcTemplateRaw)
	if err != nil {
		log.Fatal(err)
	}

	err = svcTempl.Execute(
		file,
		map[string]string{
			"name": svcName,
			"dir":  remoteDir,
			"ram":  "50M",
			"cpu":  "10%",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var rootCmd = newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
