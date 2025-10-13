# 🧾 Ansible Playbook: Datadog Node Monitoring (Docker + APM)

## 📌 Overview
This Ansible playbook automates the installation and configuration of the **Datadog Agent** for monitoring **Docker containers** and **application performance (APM)** on target nodes.  
It also deploys your `todo-app` containerized application stack for monitoring integration testing.

---

## ⚙️ Features
- Installs **Datadog Agent v7** using the **official Datadog installation script**
- Configures:
  - **Docker monitoring** (via `/var/run/docker.sock`)
  - **APM tracing** on port `8126`
  - **Logs collection** from containers
  - **Process monitoring**
- Adds the `dd-agent` user to the `docker` group for container metrics access
- Automatically restarts and enables the Datadog Agent service
- Deploys the **todo-app** containers via Docker Compose

---

## 🧩 Prerequisites

Before running this playbook, ensure:
1. **Ansible** is installed on your control node:  
   ```bash
   sudo apt install ansible -y
   ```
2. Target hosts in the `web` group are accessible via SSH  
3. **Docker** and **docker-compose** are installed on the target node(s)  
4. You have a valid **Datadog API key** and **Datadog site** (e.g., `datadoghq.com`)

---

## 📂 Playbook Tasks Overview

| Task | Description |
|------|--------------|
| **Update apt cache** | Refreshes the apt package metadata |
| **Install dependencies** | Installs required utilities like `curl` |
| **Install Datadog Agent (via install script)** | Runs Datadog’s official installation script using your API key and site |
| **Add dd-agent to docker group** | Grants the Datadog Agent access to Docker metrics |
| **Configure Datadog Agent** | Writes `/etc/datadog-agent/datadog.yaml` for APM, logs, and Docker monitoring |
| **Enable Docker integration** | Creates `/etc/datadog-agent/conf.d/docker.d/conf.yaml` |
| **Restart Datadog Agent** | Restarts and enables the Datadog service |
| **Deploy todo-app containers** | Runs `docker-compose up -d` |
| **Start todo-app stack** | Ensures containers are running |

---

## 🧰 Variables

| Variable | Description | Example |
|-----------|-------------|----------|
| `datadog_api_key` | Your unique Datadog API key | `0cfecf160dd85aa7210b66312e7571ef` |
| `datadog_site` | Datadog region site | `datadoghq.com` |
| `ansible_user` | The remote user that runs Docker | `rafsun` |

---

## 🚀 Usage

1. **Clone the repository**
   ```bash
   git clone git@github.com:rafsunl/Ansible-datadog-doc.git
   cd Ansible-datadog-doc
   ```

2. **Create an inventory file (`hosts`)**
   ```ini
   [web]
   ansible-node-1 ansible_host=192.168.0.101 ansible_user=rafsun
   ```

3. **Run the playbook**
   ```bash
   ansible-playbook -i hosts datadog_node_monitor.yml
   ```

4. **Verify installation**
   ```bash
   sudo datadog-agent status
   ```

5. **Check Datadog Dashboard**
   - Go to [https://app.datadoghq.com](https://app.datadoghq.com)
   - Navigate to **Infrastructure → Host Map**
   - Confirm the host is reporting data
   - Under **APM → Services**, check for traces from your app

---

## 📊 Expected Outcome

After successful execution:
- Datadog Agent is installed using the **official install script**
- Docker and APM integrations are active
- `todo-app` containers are running and monitored in Datadog

---

## 🔒 Security Notes
- Never hardcode your Datadog API key in playbooks  
  Use **Ansible Vault** instead:
  ```bash
  ansible-vault encrypt_string 'YOUR_API_KEY' --name 'datadog_api_key'
  ```
- Restrict SSH access to trusted hosts only

---

## 🧠 Troubleshooting

| Issue | Possible Cause | Solution |
|-------|----------------|-----------|
| `Permission denied on docker.sock` | `dd-agent` not added to docker group | Re-run the playbook or run `sudo usermod -aG docker dd-agent` |
| `Agent not reporting to Datadog` | Invalid API key or site | Check `/etc/datadog-agent/datadog.yaml` and restart agent |
| `docker-compose: command not found` | Docker Compose missing | Install it: `sudo apt install docker-compose -y` |

---

## 🧑‍💻 Author
**Md Rafsun**  
Ansible Automation & Observability Engineer  
[GitHub: @rafsunl](https://github.com/rafsunl)
