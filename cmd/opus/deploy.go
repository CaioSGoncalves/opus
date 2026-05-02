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
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy a Go service",
		Run:   runDeploy,
	}

	cmd.Flags().StringVar(&ServiceName, "name", "", "service name")
	cmd.Flags().StringVar(&TargetName, "target", "", "setup target (e.g. homelab)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("target")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) {
	target := getOpusConfigName(TargetName)

	fmt.Println("Deploy:", ServiceName, "→", target)

	// ---------------- VALIDATE BUILD ----------------

	binPath := filepath.Join("bin", ServiceName)
	if _, err := os.Stat(binPath); err != nil {
		log.Fatalf("binary not found: %s (run `make build-linux` first)", binPath)
	}

	// ---------------- PREP TEMP ----------------

	tempDir := filepath.Join(LocalTempBaseDir, ServiceName)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	tmpBin := filepath.Join(tempDir, ServiceName)

	// copy binary into temp
	mustRun("cp", binPath, tmpBin)

	// generate service file
	remoteBinPath := filepath.Join(RemoteBinBaseDir, ServiceName)
	svcPath := generateServiceFile(tempDir, ServiceName, remoteBinPath)

	// ---------------- UPLOAD ----------------

	mustRun("scp", tmpBin, target+":"+filepath.Join(RemoteTempBaseDir, ServiceName))
	mustRun("scp", svcPath, target+":"+filepath.Join(RemoteTempBaseDir, ServiceName+".service"))

	// ---------------- EXECUTE ----------------

	mustRun("ssh", "-t", target,
		fmt.Sprintf("sudo %s %s", DeployScriptPath, ServiceName),
	)

	// ---------------- CLEAN ----------------

	os.RemoveAll(tempDir)
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
