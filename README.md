# Opus 🛠️

**Build, deploy, and monitor Go services with a single tool — no YAML, no containers, no complexity.**

Opus is a unified management suite for homelabs that favors the host over the container. It handles the entire lifecycle of your Go applications using native Linux tools.

---

### Features
* **Simple Deploy:** Ship binaries directly to `systemd` via SSH.
* **Built-in Telemetry:** A lightweight worker tracks CPU, RAM, and HTTP requests.
* **Live Logs:** Stream logs directly to your terminal or web dashboard.
* **Dashboard:** A clean Web UI built with Go, Templ, and HTMX.

### Components
* **CLI:** Your interface for building and deploying.
* **Worker:** A silent agent for telemetry extraction.
* **WebApp:** Visual dashboard for services and server health.

### Tech Stack
* **Language:** Go
* **Frontend:** Templ & HTMX
* **Orchestration:** Systemd
* **Communication:** SSH / SCP
