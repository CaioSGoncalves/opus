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
	serviceName string
	remoteHost  string
)

const (
	LocalTempBaseDir     = "/tmp"
	RemoteTempBaseDir    = "/tmp"
	RemoteBinBaseDir     = "/opt/services"
	RemoteSystemdBaseDir = "/etc/systemd/system"
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
	binPath := filepath.Join("./", "bin", serviceName)
	tempDir := filepath.Join(LocalTempBaseDir, serviceName)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting deploy for:", serviceName)
	remoteBinPath := deployBinary(binPath, serviceName)
	deployService(tempDir, serviceName, remoteBinPath)

	err = os.RemoveAll(tempDir)
	if err != nil {
		log.Fatal(err)
	}
}

func deployBinary(binPath string, svcName string) string {
	remoteTempBinPath := filepath.Join(RemoteTempBaseDir, svcName)
	remoteFinalBinPath := filepath.Join(RemoteBinBaseDir, svcName)

	mustRun("scp", "-r", binPath, remoteHost+":"+remoteTempBinPath)
	sshCmd := fmt.Sprintf(
		"sudo bash -c 'mkdir -p %[1]s && mv %[2]s %[3]s && chmod +x %[3]s'",
		RemoteBinBaseDir,
		remoteTempBinPath,
		remoteFinalBinPath,
	)
	mustRun("ssh", "-t", remoteHost, sshCmd)

	return remoteFinalBinPath
}

func deployService(tempDir string, svcName string, remoteBinPath string) {
	svcPath := generateServiceFile(tempDir, svcName, remoteBinPath)
	remoteTempSvcPath := filepath.Join(RemoteTempBaseDir, svcName+".service")
	remoteFinalSvcPath := filepath.Join(RemoteSystemdBaseDir, svcName+".service")

	mustRun("scp", svcPath, remoteHost+":"+remoteTempSvcPath)

	sshCmd := fmt.Sprintf(
		"sudo bash -c '"+
			"mv -f %[2]s %[3]s && "+
			"chmod +x %[3]s && "+
			"systemctl daemon-reload && "+
			"systemctl enable %[4]s && "+
			"systemctl restart %[4]s'",
		RemoteSystemdBaseDir,
		remoteTempSvcPath,
		remoteFinalSvcPath,
		svcName,
	)
	mustRun("ssh", "-t", remoteHost, sshCmd)
}

func generateServiceFile(tempDir string, svcName string, remoteBinPath string) string {
	svcFilePath := filepath.Join(tempDir, svcName+".service")
	file, err := os.Create(svcFilePath)
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
			"dir":  RemoteBinBaseDir,
			"bin":  remoteBinPath,
			"ram":  "50M",
			"cpu":  "10%",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return svcFilePath
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

func main() {
	var rootCmd = newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
