# Datadog Agent Deployment with Ansible

## Overview

This documentation covers the automated deployment of the Datadog Agent on web servers using Ansible for comprehensive container and APM (Application Performance Monitoring) monitoring of the todo-app application.

## Architecture

The deployment sets up:
- **Datadog Agent 7**: Latest major version with full monitoring capabilities
- **Docker Container Monitoring**: Real-time container metrics and logs
- **APM (Application Performance Monitoring)**: Distributed tracing on port 8126
- **Process Monitoring**: Track system and application processes
- **Log Collection**: Automatic collection from all Docker containers

## Prerequisites

Before running the playbook, ensure:

1. **Ansible Control Node**:
   - Ansible installed (2.9 or higher recommended)
   - SSH access to target hosts configured
   - Inventory file with `web` hosts group defined

2. **Target Hosts**:
   - Ubuntu/Debian-based OS
   - Docker and Docker Compose installed
   - SSH user with sudo privileges
   - todo-app application code in `/home/<user>/todo-app`

3. **Datadog Account**:
   - Active Datadog account
   - Valid API key from your Datadog organization

## Configuration Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `datadog_api_key` | Your Datadog API key (found in Organization Settings) | `***********************` |
| `datadog_site` | Datadog site endpoint for your region | `datadoghq.com` (US1), `datadoghq.eu` (EU), `us3.datadoghq.com` (US3) |

## Playbook Structure

The playbook performs the following tasks in sequence:

### 1. System Preparation
- Updates apt package cache
- Installs curl for downloading the Datadog installation script

### 2. Datadog Agent Installation
- Downloads and executes the official Datadog Agent installation script
- Installs Agent version 7 with provided API key
- Creates `/etc/datadog-agent/datadog.yaml` configuration file

### 3. Docker Integration Setup
- Adds `dd-agent` user to the `docker` group for socket access
- Enables the agent to collect Docker metrics without permission issues

### 4. Agent Configuration
Configures the following features:

**Core Settings**:
- Binds to `0.0.0.0` for network accessibility
- Enables log collection from all containers
- Sets open file limit to 500 for log processing

**APM Configuration**:
- Enables APM receiver on port `8126`
- Ready to accept traces from instrumented applications

**Process Monitoring**:
- Enables full process collection and monitoring

**Auto-Discovery**:
- Configures Docker listener for automatic container discovery
- Enables Docker config provider with polling

### 5. Docker Monitoring
- Creates Docker integration configuration at `/etc/datadog-agent/conf.d/docker.d/conf.yaml`
- Configures connection to Docker socket at `unix://var/run/docker.sock`

### 6. Service Management
- Restarts Datadog Agent to apply all configurations
- Enables Datadog Agent to start on system boot

### 7. Application Deployment
- Deploys todo-app containers using docker-compose
- Ensures all containers are running in detached mode

## Deployment Steps

### 1. Prepare Inventory File

Create or update your Ansible inventory file:

```ini
[web]
web-server-1 ansible_host=192.168.1.10 ansible_user=ubuntu
web-server-2 ansible_host=192.168.1.11 ansible_user=ubuntu
```

### 2. Update Variables

Edit the playbook and set your Datadog credentials:

```yaml
vars:
  datadog_api_key: "your_actual_api_key_here"
  datadog_site: "datadoghq.com"  # Change based on your region
```

**Security Best Practice**: Store sensitive variables in Ansible Vault:

```bash
ansible-vault create vars/datadog_secrets.yml
```

Add to vault file:
```yaml
datadog_api_key: "your_actual_api_key_here"
```

Update playbook to include:
```yaml
vars_files:
  - vars/datadog_secrets.yml
```

### 3. Run the Playbook

Execute the playbook:

```bash
ansible-playbook -i inventory.ini datadog-setup.yml
```

With vault encryption:
```bash
ansible-playbook -i inventory.ini datadog-setup.yml --ask-vault-pass
```

### 4. Verify Installation

Check Datadog Agent status on the target host:

```bash
sudo datadog-agent status
```

Verify Docker integration:
```bash
sudo datadog-agent check docker
```

Check APM status:
```bash
sudo datadog-agent status | grep -A 10 "APM Agent"
```

## Monitoring Features Enabled

### Container Monitoring
- **Metrics**: CPU, memory, network I/O, disk I/O per container
- **Events**: Container start, stop, die, kill events
- **Labels**: Automatic collection of container labels and tags

### Log Collection
- **Automatic Collection**: All container stdout/stderr logs
- **File-based Collection**: Uses Docker log files for reliability
- **Log Processing**: Automatic parsing and tagging

### APM (Application Performance Monitoring)
- **Trace Collection**: Receives traces on port 8126
- **Service Mapping**: Automatic service dependency mapping
- **Performance Metrics**: Latency, throughput, error rates

### Process Monitoring
- **Process List**: All running processes on the host
- **Resource Usage**: CPU and memory per process
- **Process Relationships**: Parent-child process tracking

## Application Instrumentation

To send traces from your todo-app to Datadog:

### 1. Configure Environment Variables

Update your `docker-compose.yml` to include:

```yaml
services:
  app:
    environment:
      - DD_AGENT_HOST=datadog-agent  # or host IP
      - DD_TRACE_AGENT_PORT=8126
      - DD_ENV=production
      - DD_SERVICE=todo-app
      - DD_VERSION=1.0.0
```

### 2. Install Datadog Tracer

Depending on your application language:

**Node.js**:
```bash
npm install --save dd-trace
```

**Python**:
```bash
pip install ddtrace
```

**Java**:
Download `dd-java-agent.jar` and add to your container

### 3. Initialize Tracer

**Node.js** (at app entry point):
```javascript
require('dd-trace').init();
```

**Python**:
```bash
ddtrace-run python app.py
```

## Troubleshooting

### Agent Not Reporting

1. Check agent status:
```bash
sudo datadog-agent status
```

2. Verify API key:
```bash
sudo cat /etc/datadog-agent/datadog.yaml | grep api_key
```

3. Check agent logs:
```bash
sudo tail -f /var/log/datadog/agent.log
```

### Docker Metrics Missing

1. Verify dd-agent user in docker group:
```bash
groups dd-agent
```

2. Test Docker socket access:
```bash
sudo -u dd-agent docker ps
```

3. Restart agent after group change:
```bash
sudo systemctl restart datadog-agent
```

### APM Not Receiving Traces

1. Verify APM is enabled:
```bash
sudo datadog-agent status | grep -A 5 "APM Agent"
```

2. Check port 8126 is listening:
```bash
sudo netstat -tlnp | grep 8126
```

3. Verify application can reach agent:
```bash
# From application container
telnet <agent-host> 8126
```

### Logs Not Appearing

1. Check log collection is enabled:
```bash
sudo cat /etc/datadog-agent/datadog.yaml | grep logs_enabled
```

2. Verify container logs exist:
```bash
docker logs <container-name>
```

3. Check log agent status:
```bash
sudo datadog-agent status | grep -A 10 "Logs Agent"
```

## Datadog Dashboard Access

After successful deployment:

1. Log in to your Datadog account at `https://app.datadoghq.com`
2. Navigate to **Infrastructure > Containers** to view Docker metrics
3. Visit **APM > Services** to see application traces
4. Check **Logs** section for container logs
5. Go to **Infrastructure > Processes** for process monitoring

## Security Considerations

1. **API Key Protection**:
   - Never commit API keys to version control
   - Use Ansible Vault for sensitive data
   - Rotate API keys periodically

2. **Network Security**:
   - APM port 8126 should only be accessible from application containers
   - Use firewall rules to restrict access

3. **User Permissions**:
   - The `dd-agent` user has Docker socket access
   - Review and audit Docker permissions regularly

4. **Log Data**:
   - Logs may contain sensitive information
   - Configure log scrubbing rules in Datadog if needed
   - Review log retention policies

## Maintenance

### Updating Datadog Agent

```bash
sudo apt-get update
sudo apt-get install --only-upgrade datadog-agent
sudo systemctl restart datadog-agent
```

### Configuration Changes

After modifying `/etc/datadog-agent/datadog.yaml`:

```bash
sudo systemctl restart datadog-agent
sudo datadog-agent status
```

### Re-running the Playbook

The playbook is idempotent and can be safely re-run:

```bash
ansible-playbook -i inventory.ini datadog-setup.yml
```

## Additional Resources

- [Datadog Agent Documentation](https://docs.datadoghq.com/agent/)
- [Docker Integration](https://docs.datadoghq.com/integrations/docker/)
- [APM Setup Guide](https://docs.datadoghq.com/tracing/)
- [Log Collection](https://docs.datadoghq.com/logs/)
- [Ansible Datadog Role](https://github.com/DataDog/ansible-datadog)

## Support

For issues related to:
- **Ansible Playbook**: Check Ansible documentation and logs
- **Datadog Agent**: Contact Datadog support or check community forums
- **Todo-App**: Review application logs and Docker container status
