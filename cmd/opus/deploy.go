package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	ServiceName string
)

func initDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy a Go service",
		Run:   runDeploy,
	}
	deployCmd.Flags().StringVar(&ServiceName, "name", "", "service name")
	deployCmd.Flags().StringVar(&RemoteHost, "host", "", "ssh config name or user@host")
	deployCmd.MarkFlagRequired("name")
	deployCmd.MarkFlagRequired("host")

	return deployCmd
}

func runDeploy(cmd *cobra.Command, args []string) {
	binPath := filepath.Join("./", "bin", ServiceName)
	tempDir := filepath.Join(LocalTempBaseDir, ServiceName)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting deploy for:", ServiceName)
	remoteBinPath := deployBinary(binPath, ServiceName)
	deployService(tempDir, ServiceName, remoteBinPath)

	err = os.RemoveAll(tempDir)
	if err != nil {
		log.Fatal(err)
	}
}

func deployBinary(binPath string, svcName string) string {
	remoteTempBinPath := filepath.Join(RemoteTempBaseDir, svcName)
	remoteFinalBinPath := filepath.Join(RemoteBinBaseDir, svcName)

	mustRun("scp", "-r", binPath, RemoteHost+":"+remoteTempBinPath)
	sshCmd := fmt.Sprintf(
		"sudo bash -c 'mkdir -p %[1]s && mv %[2]s %[3]s && chmod +x %[3]s'",
		RemoteBinBaseDir,
		remoteTempBinPath,
		remoteFinalBinPath,
	)
	mustRun("ssh", "-t", RemoteHost, sshCmd)

	return remoteFinalBinPath
}

func deployService(tempDir string, svcName string, remoteBinPath string) {
	svcPath := generateServiceFile(tempDir, svcName, remoteBinPath)
	remoteTempSvcPath := filepath.Join(RemoteTempBaseDir, svcName+".service")
	remoteFinalSvcPath := filepath.Join(RemoteSystemdBaseDir, svcName+".service")

	mustRun("scp", svcPath, RemoteHost+":"+remoteTempSvcPath)

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
	mustRun("ssh", "-t", RemoteHost, sshCmd)
}

func generateServiceFile(tempDir string, svcName string, remoteBinPath string) string {
	svcFilePath := filepath.Join(tempDir, svcName+".service")
	file, err := os.Create(svcFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	svcTempl, err := template.New("service").Parse(SvcTemplateRaw)
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
