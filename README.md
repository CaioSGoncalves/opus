# Opus 🛠️

**Fast, simple deploys for your homelab — binaries with systemd or containers with k3s.**

Opus is designed for homelabs and small servers, focusing on direct deployments (systemd or k3s) without unnecessary complexity. It provides a clean developer experience while remaining flexible for future expansion.

---

## Example usage

### 1. Setup server access

```bash
opus setup init --target homelab --host user@192.168.1.10
```
### 2. Deploy a service (systemd)
```bash
opus deploy --name domus --host homelab
```
### 3. Setup k3s + kubectl
```bash
opus setup k3s --target homelab
```
### 4. Use the cluster
```bash
opus use homelab
kubectl get nodes
```
### 5. SSH into server
```bash
ssh homelab_opus
```


## Features
- Simple Deploy: Ship Go binaries via SSH using systemd
- K3s Support: Install and configure Kubernetes (k3s) easily
- Context Switching: Seamless kubectl switching via opus use
- Minimal Setup: No YAML required for basic usage

## Concepts
- Targets are isolated (homelab, server1, etc.)
- Each environment has its own kubeconfig
- No global kubeconfig mutation by default
- Explicit over implicit

## Limitations
- No rollback or versioning
- No secrets/config management
- Assumes simple environments (single node or small setups)

## Tech Stack
- Language: Go
- Orchestration: Systemd / K3s
- Communication: SSH / SCP

## WIP
- Built-in telemetry (CPU, RAM, HTTP metrics)
- Background worker for metrics collection
- Live logs streaming
- Web dashboard for services and server health
