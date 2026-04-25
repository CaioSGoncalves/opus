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

var (
	RemoteHost string

	SSHConfigPath    = filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	DeployScriptPath = "/usr/local/bin/opus-deploy"
)

// ---------------- ROOT SETUP ----------------

func initSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Manage setup lifecycle",
	}

	cmd.AddCommand(initSetupInitCmd())
	cmd.AddCommand(initSetupClearCmd())

	return cmd
}

// ---------------- INIT ----------------

func initSetupInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Setup SSH + server for deploy",
		Run:   runSetupInit,
	}

	cmd.Flags().StringVar(&TargetName, "target", "", "name (e.g. homelab)")
	cmd.Flags().StringVar(&RemoteHost, "host", "", "user@host")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("host")

	return cmd
}

func runSetupInit(cmd *cobra.Command, args []string) {
	target := TargetName + "_opus"
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519_"+target)

	fmt.Println("Setup:", target, "->", RemoteHost)

	ensureKey(keyPath, target)
	writeSSHConfig(target, RemoteHost, keyPath)
	copySSHKey(target)

	installDeployScript(target)
	ensureSudoers(target)

	fmt.Println("Setup complete ✔")
}

// ---------------- CLEAR ----------------

func initSetupClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove setup",
		Run:   runSetupClear,
	}

	cmd.Flags().StringVar(&TargetName, "target", "", "name (e.g. homelab)")
	cmd.MarkFlagRequired("target")

	return cmd
}

func runSetupClear(cmd *cobra.Command, args []string) {
	target := TargetName + "_opus"
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519_"+target)

	fmt.Println("Clearing setup for:", target)

	// order matters
	removeRemote(target)
	removeSSHConfig(target)
	removeSSHKey(keyPath)

	fmt.Println("Clear complete ✔")
}

// ---------------- SSH ----------------

func ensureKey(path string, target string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Generating SSH key...")
		mustRun("ssh-keygen",
			"-t", "ed25519",
			"-f", path,
			"-C", target,
			"-N", "",
		)
	}
}

func writeSSHConfig(target, host, keyPath string) {
	data, _ := os.ReadFile(SSHConfigPath)
	content := string(data)

	for line := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(line) == "Host "+target {
			fmt.Println("SSH config already exists, skipping")
			return
		}
	}

	user := extractUser(host)
	hostname := extractHost(host)

	block := fmt.Sprintf(
		"\nHost %[1]s\n"+
			"    HostName %[2]s\n"+
			"    User %[3]s\n"+
			"    IdentityFile %[4]s\n"+
			"    AddKeysToAgent yes\n"+
			"    UseKeychain yes\n",
		target, hostname, user, keyPath,
	)

	f, err := os.OpenFile(SSHConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.WriteString(block); err != nil {
		log.Fatal(err)
	}
}

func copySSHKey(target string) {
	fmt.Println("Copying SSH key (password once)...")
	mustRun("ssh-copy-id", target)
}

// ---------------- REMOTE ----------------

func installDeployScript(target string) {
	fmt.Println("Installing deploy script...")

	script := `#!/bin/bash
set -e

NAME=$1

mv -f /tmp/$NAME /opt/services/$NAME
chmod +x /opt/services/$NAME

mv -f /tmp/$NAME.service /etc/systemd/system/$NAME.service
chmod 644 /etc/systemd/system/$NAME.service

systemctl daemon-reload
systemctl enable $NAME
systemctl restart $NAME
`

	cmd := fmt.Sprintf(
		"echo '%[1]s' | sudo tee %[2]s > /dev/null && sudo chmod +x %[2]s",
		strings.ReplaceAll(script, "'", "'\\''"),
		DeployScriptPath,
	)

	mustRun("ssh", "-t", target, cmd)
}

func ensureSudoers(target string) {
	fmt.Println("Ensuring sudoers...")

	user := extractUser(RemoteHost)
	line := fmt.Sprintf("%[1]s ALL=(ALL) NOPASSWD: %[2]s", user, DeployScriptPath)

	check := fmt.Sprintf("sudo grep -q '%s' /etc/sudoers", line)
	if exec.Command("ssh", target, check).Run() == nil {
		return
	}

	fmt.Println("Adding sudoers rule...")

	cmd := fmt.Sprintf(
		"echo '%s' | sudo EDITOR='tee -a' visudo",
		line,
	)

	mustRun("ssh", "-t", target, cmd)
}

// ---------------- CLEAR HELPERS ----------------

func removeSSHConfig(target string) {
	data, err := os.ReadFile(SSHConfigPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	skip := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "Host "+target {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(line, "Host ") {
			skip = false
		}
		if !skip {
			out = append(out, line)
		}
	}

	os.WriteFile(SSHConfigPath, []byte(strings.Join(out, "\n")), 0600)
}

func removeSSHKey(keyPath string) {
	os.Remove(keyPath)
	os.Remove(keyPath + ".pub")
}

func removeRemote(target string) {
	fmt.Println("Removing remote setup...")

	cmd := fmt.Sprintf(
		"sudo rm -f %[1]s && sudo sed -i '\\|%[1]s|d' /etc/sudoers",
		DeployScriptPath,
	)

	mustRun("ssh", "-t", target, cmd)
}

// ---------------- UTILS ----------------

func extractUser(h string) string {
	parts := strings.Split(h, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return "root"
}

func extractHost(h string) string {
	parts := strings.Split(h, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return h
}
