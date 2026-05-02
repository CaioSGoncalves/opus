package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func initSetupK3sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k3s",
		Short: "Setup k3s and kubectl access",
		Run:   runSetupK3s,
	}

	cmd.Flags().StringVar(&TargetName, "target", "", "target (e.g. homelab)")
	cmd.MarkFlagRequired("target")

	return cmd
}

func runSetupK3s(cmd *cobra.Command, args []string) {
	target := getOpusConfigName(TargetName)

	fmt.Println("Setup k3s:", target)

	// ---------------- CHECK kubectl ----------------

	if !runLocalCheck("kubectl version --client") {
		log.Fatal("kubectl not found. Install kubectl first.")
	}

	// ---------------- INSTALL K3S ----------------

	if runSSHCheck(target, "command -v k3s") {
		fmt.Println("k3s already installed → skipping")
	} else {
		fmt.Println("Installing k3s...")
		mustRun("ssh", "-t", target, "curl -sfL https://get.k3s.io | sh -")
	}

	// ---------------- PATHS ----------------

	home := os.Getenv("HOME")
	kubeDir := filepath.Join(home, ".kube")
	os.MkdirAll(kubeDir, 0755)

	localConfig := filepath.Join(kubeDir, "config-opus-"+TargetName)

	// ---------------- FETCH KUBECONFIG ----------------

	fmt.Println("Fetching kubeconfig...")

	mustRun("ssh", "-t", target, `
sudo cat /etc/rancher/k3s/k3s.yaml > ~/kubeconfig-opus.yaml
sudo chown $USER:$USER ~/kubeconfig-opus.yaml
`)

	mustRun("scp", target+":~/kubeconfig-opus.yaml", localConfig)

	// ---------------- FIX SERVER ----------------

	data, err := os.ReadFile(localConfig)
	if err != nil {
		log.Fatal(err)
	}

	host := extractHostFromTarget(target)
	fixed := strings.ReplaceAll(string(data), "127.0.0.1", host)

	if err := os.WriteFile(localConfig, []byte(fixed), 0600); err != nil {
		log.Fatal(err)
	}

	// ---------------- INSTALL SHELL FUNCTION ----------------

	installShellFunction()
	fmt.Println("Restart terminal or run: source ~/.zshrc (or ~/.bashrc)")
	fmt.Println("Setup complete ✔")
}

// ---------------- SHELL FUNCTION ----------------

func installShellFunction() {
	home := os.Getenv("HOME")

	shellFiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
	}

	function := `
# opus shell integration
opus() {
  if [ "$1" = "use" ]; then
    shift
    eval "$(command opus use "$@")"
  else
    command opus "$@"
  fi
}
`

	for _, file := range shellFiles {
		data, _ := os.ReadFile(file)
		if strings.Contains(string(data), "opus shell integration") {
			continue // already installed
		}

		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		defer f.Close()

		f.WriteString("\n" + function + "\n")

		fmt.Println("Shell integration added to:", file)
	}

	fmt.Println("Restart your shell or run: source ~/.zshrc")
}

// ---------------- HELPERS ----------------

func runSSHCheck(target, cmd string) bool {
	return exec.Command("ssh", target, cmd).Run() == nil
}

func runLocalCheck(cmd string) bool {
	parts := strings.Split(cmd, " ")
	return exec.Command(parts[0], parts[1:]...).Run() == nil
}

func extractHostFromTarget(target string) string {
	out, err := exec.Command("ssh", "-G", target).Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "hostname ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "hostname "))
		}
	}

	return ""
}
