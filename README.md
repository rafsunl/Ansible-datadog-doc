# Complete Todo App Deployment with Datadog Monitoring

## Overview

This guide takes you from scratch to a fully deployed todo-app with Datadog monitoring:

**End Result:** Todo app running with full monitoring (containers, APM, logs, metrics)

---

## Step 1: Setup Passwordless SSH Authentication

### On Your Control Machine (where you run Ansible)

```bash
# Generate SSH key if you don't have one
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
# Press Enter for all prompts (use default location, no passphrase)

# Copy your SSH key to the target server
ssh-copy-id [user]@YOUR_SERVER_IP

# Test passwordless login
ssh [user]@YOUR_SERVER_IP
# Should login without asking for password
exit
```

Replace `[user]` with your actual username on the target server.

**Verify:** You can SSH without entering a password ✅

---

## Step 2: Install Ansible

### On Your Control Machine

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ansible -y
ansible --version
```
**Verify:** Should show Ansible version 2.9+ ✅

---

## Step 3: Project Setup

### Create Project Structure

```bash
mkdir -p ~/ansible-todo-deployment
cd ~/ansible-todo-deployment
mkdir -p vars playbooks
```

### Create Inventory File

Create `inventory.ini`:
```ini
[web]
ansible-node-1 ansible_host=YOUR_SERVER_IP ansible_user=[USER]

[web:vars]
ansible_python_interpreter=/usr/bin/python3
```

Replace `YOUR_SERVER_IP` with your actual server IP.

---

## Step 4: Install Docker on Target Server

### Create Docker Installation Playbook

Create `playbooks/install-docker.yml`:

```yaml
---
- name: Install Docker and Docker Compose
  hosts: web
  become: yes
  tasks:
    - name: Update apt cache
      apt:
        update_cache: yes

    - name: Install required packages
      apt:
        name:
          - apt-transport-https
          - ca-certificates
          - curl
          - gnupg
          - lsb-release
        state: present

    - name: Add Docker GPG key
      apt_key:
        url: https://download.docker.com/linux/ubuntu/gpg
        state: present

    - name: Add Docker repository
      apt_repository:
        repo: "deb [arch=amd64] https://download.docker.com/linux/ubuntu {{ ansible_distribution_release }} stable"
        state: present

    - name: Install Docker
      apt:
        name:
          - docker-ce
          - docker-ce-cli
          - containerd.io
        state: present
        update_cache: yes

    - name: Add user to docker group
      user:
        name: "{{ ansible_user }}"
        groups: docker
        append: yes

    - name: Install Docker Compose
      get_url:
        url: "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-linux-x86_64"
        dest: /usr/local/bin/docker-compose
        mode: '0755'

    - name: Enable Docker service
      systemd:
        name: docker
        enabled: yes
        state: started
```

### Run Docker Installation

```bash
ansible-playbook -i inventory.ini playbooks/install-docker.yml
```

**Verify:** SSH to server and run `docker --version` ✅

---

## Step 5: Deploy Todo App Code

### Transfer Application Code to Server

```bash
# On your control machine, create the todo-app directory structure
mkdir -p todo-app/{frontend,backend}

# Copy your application code to todo-app/
# (Make sure you have frontend and backend folders with their code)

# Transfer to server
scp -r todo-app rafsun@YOUR_SERVER_IP:/home/[user]/
```

### Create docker-compose.yml

SSH to your server and create `/home/[user]/todo-app/docker-compose.yml`:

```yaml
version: '3.8'
services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    depends_on:
      - backend
    labels:
      com.datadoghq.ad.logs: '[{"source": "frontend", "service": "todo-frontend"}]'

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=todo
      - DD_AGENT_HOST=172.17.0.1
      - DD_ENV=production
      - DD_SERVICE=todo-backend
      - DD_VERSION=1.0
    depends_on:
      - db
    labels:
      com.datadoghq.ad.logs: '[{"source": "backend", "service": "todo-backend"}]'

  db:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: todo
    ports:
      - "5432:5432"
    volumes:
      - db-data:/var/lib/postgresql/data
    labels:
      com.datadoghq.ad.logs: '[{"source": "postgresql", "service": "todo-db"}]'

volumes:
  db-data:
```

**Note:** The `DD_AGENT_HOST=172.17.0.1` points to Docker host where Datadog agent will run.

---

## Step 6: Setup Datadog Credentials

### Create Encrypted Vault

```bash
cd ~/ansible-todo-deployment
ansible-vault create vars/datadog_secrets.yml
```

Enter a password and add:
```yaml
datadog_api_key: "YOUR_DATADOG_API_KEY"
datadog_site: "datadoghq.com"
```

Get your API key from: https://app.datadoghq.com/organization-settings/api-keys

**Save and exit** 
---

## Step 7: Deploy Datadog + Start Application

### Create Main Deployment Playbook on the main host

Create `playbooks/deploy-with-datadog.yml`:

```yaml
---
- name: Install Datadog Agent and Deploy Todo App
  hosts: web
  become: yes
  vars_files:
    - ../vars/datadog_secrets.yml
  tasks:
    - name: Update apt cache
      apt:
        update_cache: yes

    - name: Install curl
      apt:
        name: curl
        state: present

    - name: Install Datadog Agent
      shell: |
        DD_AGENT_MAJOR_VERSION=7 DD_API_KEY={{ datadog_api_key }} DD_SITE={{ datadog_site }} bash -c "$(curl -L https://s3.amazonaws.com/dd-agent/scripts/install_script.sh)"
      args:
        creates: /etc/datadog-agent/datadog.yaml

    - name: Add dd-agent to docker group
      user:
        name: dd-agent
        groups: docker
        append: yes

    - name: Configure Datadog Agent
      copy:
        dest: /etc/datadog-agent/datadog.yaml
        content: |
          api_key: {{ datadog_api_key }}
          site: {{ datadog_site }}
          bind_host: 0.0.0.0
          logs_enabled: true
          logs_config:
            container_collect_all: true
          apm_config:
            enabled: true
            receiver_port: 8126
          process_config:
            enabled: true
          listeners:
            - name: docker
          config_providers:
            - name: docker
              polling: true

    - name: Enable Docker integration
      copy:
        dest: /etc/datadog-agent/conf.d/docker.d/conf.yaml
        content: |
          init_config:
          instances:
            - url: "unix://var/run/docker.sock"

    - name: Restart Datadog Agent
      systemd:
        name: datadog-agent
        state: restarted
        enabled: yes

    - name: Deploy todo-app with docker-compose
      shell: docker-compose up -d --build
      args:
        chdir: /home/{{ ansible_user }}/todo-app
      become_user: "{{ ansible_user }}"
```

### Run the Deployment

```bash
cd ~/ansible-todo-deployment
ansible-playbook -i inventory.ini playbooks/deploy-with-datadog.yml --ask-vault-pass
```

Enter your vault password when prompted.

---

## Step 8: Verify Everything Works

### Check Application

```bash
# Frontend
curl http://YOUR_SERVER_IP:3000

# Backend
curl http://YOUR_SERVER_IP:8080

# Check containers
ssh rafsun@YOUR_SERVER_IP
docker ps
```

You should see 3 containers running:
- `todo-app-frontend-1`
- `todo-app-backend-1`
- `todo-app-db-1`

### Check Datadog Agent

```bash
ssh [user]@YOUR_SERVER_IP
sudo datadog-agent status
```

Look for:
- ✅ Status: Running
- ✅ Docker Check: OK
- ✅ APM Agent: Running (port 8126)
- ✅ Logs Agent: Running

### View in Datadog Dashboard

1. Go to https://app.datadoghq.com
2. **Containers**: Infrastructure → Containers
3. **APM**: APM → Services (look for `todo-backend`)
4. **Logs**: Logs → Search

---
---

## Complete Deployment Checklist

- [ ] Passwordless SSH working
- [ ] Ansible installed
- [ ] Inventory file created
- [ ] Docker installed on server
- [ ] Todo app code on server
- [ ] docker-compose.yml created
- [ ] Datadog vault file created
- [ ] Deployment playbook run successfully
- [ ] Containers running (docker ps shows 3 containers)
- [ ] Datadog agent status OK
- [ ] App accessible (frontend on :3000, backend on :8080)
- [ ] Metrics visible in Datadog dashboard

---

## Resources

- [Ansible Docs](https://docs.ansible.com)
- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Datadog Integration Docs](https://docs.datadoghq.com/integrations/docker/)
