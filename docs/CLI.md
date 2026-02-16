# InfraAudit CLI Documentation

Command-line interface for the InfraAudit Cloud Infrastructure Auditing and Security Platform.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Global Flags](#global-flags)
- [Output Formats](#output-formats)
- [Commands](#commands)
  - [auth](#auth) - Authentication
  - [config](#config) - CLI configuration
  - [status](#status) - Dashboard summary
  - [provider](#provider) - Cloud providers
  - [resource](#resource) - Cloud resources
  - [drift](#drift) - Security drifts
  - [alert](#alert) - Alerts
  - [vulnerability](#vulnerability) - Vulnerabilities
  - [cost](#cost) - Cost analytics
  - [compliance](#compliance) - Compliance frameworks
  - [kubernetes](#kubernetes) - Kubernetes clusters
  - [iac](#iac) - Infrastructure as Code
  - [job](#job) - Scheduled jobs
  - [remediation](#remediation) - Remediation actions
  - [recommendation](#recommendation) - AI recommendations
  - [notification](#notification) - Notifications
  - [webhook](#webhook) - Webhooks
- [Shell Completion](#shell-completion)
- [Environment Variables](#environment-variables)
- [Examples](#examples)

---

## Installation

### From Source

Requires Go 1.24+.

```bash
# Clone the repository
git clone https://github.com/pratik-mahalle/infraudit-go.git
cd infraudit-go

# Build the CLI binary
make build-cli

# Binary is at ./bin/infraaudit
./bin/infraaudit --help
```

### Go Install

```bash
go install github.com/pratik-mahalle/infraudit/cmd/cli@latest
```

The binary will be installed to `$GOPATH/bin/` (or `$HOME/go/bin/` by default). Make sure this directory is in your `PATH`.

### Verify Installation

```bash
infraaudit --help
```

---

## Quick Start

```bash
# 1. Run first-time setup
infraaudit config init
# > Enter server URL: http://localhost:8080
# > Default output format: table

# 2. Login
infraaudit auth login
# > Email: user@example.com
# > Password: ********
# > Logged in as user@example.com

# 3. Check dashboard
infraaudit status

# 4. Connect a cloud provider
infraaudit provider connect aws

# 5. List resources
infraaudit resource list

# 6. Run drift detection
infraaudit drift detect
```

---

## Configuration

Configuration is stored at `~/.infraaudit/config.yaml`. It is created automatically during `config init` or `auth login`.

### Config File Structure

```yaml
server_url: http://localhost:8080
output: table
auth:
  token: <jwt-token>
  refresh_token: <refresh-token>
  email: user@example.com
```

The config file is created with `0600` permissions (owner read/write only) to protect credentials.

### Managing Config

```bash
# Interactive setup
infraaudit config init

# Set individual values
infraaudit config set server_url https://api.infraaudit.dev
infraaudit config set output json

# Read a value
infraaudit config get server_url

# List all values
infraaudit config list
```

---

## Global Flags

These flags are available on every command:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--server` | | (from config) | Override the server URL for this request |
| `--output` | `-o` | `table` | Output format: `table`, `json`, `yaml` |
| `--config` | | `~/.infraaudit/config.yaml` | Path to config file |
| `--no-color` | | `false` | Disable colored output |
| `--help` | `-h` | | Show help for any command |

### Examples

```bash
# Use a different server
infraaudit --server https://staging.infraaudit.dev status

# Get JSON output
infraaudit resource list -o json

# Get YAML output
infraaudit drift list -o yaml

# Use a custom config file
infraaudit --config /path/to/config.yaml status
```

---

## Output Formats

### Table (default)

Human-readable table with aligned columns:

```
ID  NAME            TYPE     REGION      STATUS         PROVIDER
--  ----            ----     ------      ------         --------
1   web-server-01   ec2      us-east-1   [+] active     1
2   api-gateway     lambda   us-west-2   [+] active     1
3   data-bucket     s3       us-east-1   [+] active     1
```

### JSON

Machine-readable JSON (use with `jq` for filtering):

```bash
infraaudit resource list -o json | jq '.[].name'
```

### YAML

YAML output:

```bash
infraaudit drift get 1 -o yaml
```

---

## Commands

### auth

Authentication commands. Credentials are stored securely in `~/.infraaudit/config.yaml`.

#### `auth login`

Login with email and password. If flags are not provided, prompts interactively (password input is hidden).

```bash
# Interactive
infraaudit auth login

# Non-interactive
infraaudit auth login --email user@example.com --password mypassword
```

| Flag | Description |
|------|-------------|
| `--email` | Email address |
| `--password` | Password |

#### `auth register`

Create a new account.

```bash
# Interactive
infraaudit auth register

# Non-interactive
infraaudit auth register --email user@example.com --name "John Doe" --password mypassword
```

| Flag | Description |
|------|-------------|
| `--email` | Email address |
| `--name` | Full name |
| `--password` | Password |

#### `auth logout`

Clear stored credentials from config.

```bash
infraaudit auth logout
```

#### `auth whoami`

Display current authenticated user info.

```bash
infraaudit auth whoami
```

---

### config

Manage CLI configuration. See [Configuration](#configuration) section for details.

| Subcommand | Description |
|------------|-------------|
| `config init` | Interactive first-time setup |
| `config set <key> <value>` | Set a config value |
| `config get <key>` | Get a config value |
| `config list` | Show all config values |

---

### status

Show a dashboard summary of your InfraAudit account.

```bash
infraaudit status
```

Output:

```
InfraAudit Dashboard
========================================
  Providers:     1 connected (1 total)
  Resources:     47 synced
  Drifts:        3 detected (1 critical)
  Alerts:        5 open (2 high severity)
```

---

### provider

Manage cloud provider connections (AWS, GCP, Azure).

#### `provider list`

List all connected providers.

```bash
infraaudit provider list
```

#### `provider connect <aws|gcp|azure>`

Connect a new cloud provider. Prompts interactively for credentials.

```bash
# Connect AWS
infraaudit provider connect aws
# > AWS Access Key ID: AKIA...
# > AWS Secret Access Key: ********
# > AWS Region [us-east-1]: us-west-2

# Connect GCP
infraaudit provider connect gcp
# > GCP Project ID: my-project
# > Path to service account JSON: /path/to/key.json

# Connect Azure
infraaudit provider connect azure
# > Azure Tenant ID: ...
# > Azure Client ID: ...
# > Azure Client Secret: ********
# > Azure Subscription ID: ...
```

#### `provider sync <provider-id>`

Trigger resource sync from a provider.

```bash
infraaudit provider sync 1
# Syncing resources...
# Sync complete: 47 found, 12 created, 35 updated
```

#### `provider disconnect <provider-id>`

Disconnect a provider.

```bash
infraaudit provider disconnect 1
```

#### `provider status`

Show sync status for all providers.

```bash
infraaudit provider status
```

---

### resource

Manage cloud resources discovered by connected providers.

#### `resource list`

List resources with optional filters.

```bash
# List all
infraaudit resource list

# Filter by provider
infraaudit resource list --provider 1

# Filter by type and region
infraaudit resource list --type ec2 --region us-east-1

# JSON output
infraaudit resource list -o json
```

| Flag | Description |
|------|-------------|
| `--provider` | Filter by provider ID |
| `--type` | Filter by resource type |
| `--region` | Filter by region |
| `--status` | Filter by status |

#### `resource get <id>`

Show detailed information about a resource.

```bash
infraaudit resource get 42
```

#### `resource delete <id>`

Delete a resource.

```bash
infraaudit resource delete 42
```

---

### drift

Detect and manage infrastructure configuration drifts.

#### `drift list`

List detected drifts with optional filters.

```bash
infraaudit drift list
infraaudit drift list --severity critical
infraaudit drift list --status detected --type security
```

| Flag | Description |
|------|-------------|
| `--severity` | Filter: `critical`, `high`, `medium`, `low` |
| `--status` | Filter: `detected`, `investigating`, `resolved` |
| `--type` | Filter: `configuration`, `security`, `compliance` |

#### `drift get <id>`

Show drift details.

```bash
infraaudit drift get 5
```

#### `drift detect`

Trigger a drift detection scan across all resources.

```bash
infraaudit drift detect
```

#### `drift summary`

Show an aggregate summary of drift status.

```bash
infraaudit drift summary
```

#### `drift resolve <id>`

Mark a drift as resolved.

```bash
infraaudit drift resolve 5
```

---

### alert

Manage security and operational alerts.

#### `alert list`

List alerts with optional filters.

```bash
infraaudit alert list
infraaudit alert list --severity high --status open
infraaudit alert list --type security
```

| Flag | Description |
|------|-------------|
| `--severity` | Filter: `critical`, `high`, `medium`, `low` |
| `--status` | Filter: `open`, `acknowledged`, `resolved` |
| `--type` | Filter: `security`, `compliance`, `performance` |

#### `alert get <id>`

Show alert details.

```bash
infraaudit alert get 12
```

#### `alert summary`

Show aggregate alert summary.

```bash
infraaudit alert summary
```

#### `alert acknowledge <id>`

Acknowledge an alert.

```bash
infraaudit alert acknowledge 12
```

#### `alert resolve <id>`

Resolve an alert.

```bash
infraaudit alert resolve 12
```

---

### vulnerability

Manage vulnerability scanning and results. **Alias:** `vuln`

#### `vulnerability list`

List vulnerabilities with optional filters.

```bash
infraaudit vuln list
infraaudit vuln list --severity critical
infraaudit vuln list --status open
```

| Flag | Description |
|------|-------------|
| `--severity` | Filter by severity |
| `--status` | Filter by status |

#### `vulnerability get <id>`

Show vulnerability details.

```bash
infraaudit vuln get 7
```

#### `vulnerability scan`

Trigger a vulnerability scan.

```bash
# Scan all resources
infraaudit vuln scan

# Scan a specific resource
infraaudit vuln scan --resource 42
```

| Flag | Description |
|------|-------------|
| `--resource` | Scan a specific resource ID |

#### `vulnerability summary`

Show vulnerability summary statistics.

```bash
infraaudit vuln summary
```

#### `vulnerability top`

Show the most critical vulnerabilities.

```bash
infraaudit vuln top
```

---

### cost

Cloud cost analytics, forecasting, and optimization.

#### `cost overview`

Show cost overview across all providers.

```bash
infraaudit cost overview
```

#### `cost trends`

Show cost trends over time.

```bash
infraaudit cost trends
infraaudit cost trends --provider aws --period 30d
```

| Flag | Description |
|------|-------------|
| `--provider` | Filter by provider |
| `--period` | Time period: `7d`, `30d`, `90d` |

#### `cost forecast`

Show cost forecast.

```bash
infraaudit cost forecast
infraaudit cost forecast --provider aws --days 90
```

| Flag | Description |
|------|-------------|
| `--provider` | Filter by provider |
| `--days` | Forecast horizon: `30`, `60`, `90` |

#### `cost sync`

Sync cost data from cloud providers.

```bash
infraaudit cost sync
infraaudit cost sync --provider aws
```

| Flag | Description |
|------|-------------|
| `--provider` | Sync specific provider |

#### `cost anomalies`

List detected cost anomalies.

```bash
infraaudit cost anomalies
```

#### `cost detect-anomalies`

Trigger cost anomaly detection.

```bash
infraaudit cost detect-anomalies
```

#### `cost optimizations`

List cost optimization opportunities.

```bash
infraaudit cost optimizations
```

#### `cost savings`

Show potential cost savings.

```bash
infraaudit cost savings
```

---

### compliance

Compliance framework management and assessment.

#### `compliance overview`

Show overall compliance posture.

```bash
infraaudit compliance overview
```

#### `compliance frameworks`

List available compliance frameworks.

```bash
infraaudit compliance frameworks
```

#### `compliance framework <id>`

Show details of a specific framework.

```bash
infraaudit compliance framework cis-aws
```

#### `compliance enable <id>`

Enable a compliance framework.

```bash
infraaudit compliance enable cis-aws
```

#### `compliance disable <id>`

Disable a compliance framework.

```bash
infraaudit compliance disable cis-aws
```

#### `compliance assess`

Run a compliance assessment.

```bash
# Assess all enabled frameworks
infraaudit compliance assess

# Assess a specific framework
infraaudit compliance assess --framework cis-aws
```

| Flag | Description |
|------|-------------|
| `--framework` | Framework ID to assess |

#### `compliance assessments`

List past assessment results.

```bash
infraaudit compliance assessments
```

#### `compliance export <assessment-id>`

Export an assessment report.

```bash
infraaudit compliance export 3
```

#### `compliance failing-controls`

Show currently failing controls.

```bash
infraaudit compliance failing-controls
```

---

### kubernetes

Kubernetes cluster management and monitoring. **Alias:** `k8s`

#### `kubernetes clusters`

List registered Kubernetes clusters.

```bash
infraaudit k8s clusters
```

#### `kubernetes register`

Register a new Kubernetes cluster.

```bash
infraaudit k8s register --name production --kubeconfig ~/.kube/config
```

| Flag | Description |
|------|-------------|
| `--name` | Cluster name |
| `--kubeconfig` | Path to kubeconfig file |

#### `kubernetes delete <id>`

Delete a registered cluster.

```bash
infraaudit k8s delete 1
```

#### `kubernetes sync <id>`

Sync resources from a cluster.

```bash
infraaudit k8s sync 1
```

#### `kubernetes namespaces <cluster-id>`

List namespaces in a cluster.

```bash
infraaudit k8s namespaces 1
```

#### `kubernetes deployments <cluster-id>`

List deployments in a cluster.

```bash
infraaudit k8s deployments 1
```

#### `kubernetes pods <cluster-id>`

List pods in a cluster.

```bash
infraaudit k8s pods 1
```

#### `kubernetes services <cluster-id>`

List services in a cluster.

```bash
infraaudit k8s services 1
```

#### `kubernetes stats`

Show aggregate cluster statistics.

```bash
infraaudit k8s stats
```

---

### iac

Infrastructure as Code management. Supports Terraform, CloudFormation, and Kubernetes manifests.

#### `iac upload <file>`

Upload an IaC file for analysis.

```bash
infraaudit iac upload main.tf
infraaudit iac upload cloudformation.yaml
infraaudit iac upload k8s-deployment.yaml
```

#### `iac definitions`

List uploaded IaC definitions.

```bash
infraaudit iac definitions
```

#### `iac detect-drift`

Detect drift between IaC definitions and actual infrastructure.

```bash
infraaudit iac detect-drift
```

#### `iac drifts`

List detected IaC drifts.

```bash
infraaudit iac drifts
```

#### `iac drift-summary`

Show IaC drift summary.

```bash
infraaudit iac drift-summary
```

---

### job

Manage scheduled automation jobs.

#### `job list`

List all scheduled jobs.

```bash
infraaudit job list
```

#### `job create`

Create a new scheduled job. Prompts interactively if flags are not provided.

```bash
infraaudit job create --name "Daily Drift Scan" --type drift_detection --schedule "0 8 * * *"
```

| Flag | Description |
|------|-------------|
| `--name` | Job name |
| `--type` | Job type (see `job types`) |
| `--schedule` | Cron schedule expression |

#### `job get <id>`

Show job details.

```bash
infraaudit job get 1
```

#### `job delete <id>`

Delete a job.

```bash
infraaudit job delete 1
```

#### `job run <id>`

Trigger immediate job execution.

```bash
infraaudit job run 1
```

#### `job executions <job-id>`

List execution history for a job.

```bash
infraaudit job executions 1
```

#### `job types`

List available job types.

```bash
infraaudit job types
```

---

### remediation

Manage automated remediation actions.

#### `remediation summary`

Show remediation summary.

```bash
infraaudit remediation summary
```

#### `remediation pending`

List pending remediation approvals.

```bash
infraaudit remediation pending
```

#### `remediation suggest-drift <drift-id>`

Generate AI-powered remediation suggestion for a drift.

```bash
infraaudit remediation suggest-drift 5
```

#### `remediation suggest-vuln <vulnerability-id>`

Generate AI-powered remediation suggestion for a vulnerability.

```bash
infraaudit remediation suggest-vuln 7
```

#### `remediation approve <action-id>`

Approve a pending remediation action.

```bash
infraaudit remediation approve 3
```

#### `remediation execute <action-id>`

Execute an approved remediation action.

```bash
infraaudit remediation execute 3
```

#### `remediation rollback <action-id>`

Rollback a previously executed remediation.

```bash
infraaudit remediation rollback 3
```

---

### recommendation

AI-powered optimization recommendations. **Alias:** `rec`

#### `recommendation list`

List recommendations with optional filters.

```bash
infraaudit rec list
infraaudit rec list --type cost --priority high
infraaudit rec list --status pending
```

| Flag | Description |
|------|-------------|
| `--type` | Filter: `cost`, `performance`, `security` |
| `--priority` | Filter: `high`, `medium`, `low` |
| `--status` | Filter: `pending`, `applied`, `dismissed` |

#### `recommendation get <id>`

Show recommendation details.

```bash
infraaudit rec get 10
```

#### `recommendation generate`

Trigger AI recommendation generation.

```bash
infraaudit rec generate
```

#### `recommendation savings`

Show total potential savings across all recommendations.

```bash
infraaudit rec savings
```

#### `recommendation apply <id>`

Mark a recommendation as applied.

```bash
infraaudit rec apply 10
```

#### `recommendation dismiss <id>`

Dismiss a recommendation.

```bash
infraaudit rec dismiss 10
```

---

### notification

Manage notification preferences and history. **Alias:** `notif`

#### `notification preferences`

Show current notification preferences.

```bash
infraaudit notif preferences
```

#### `notification update <channel>`

Update notification preference for a channel.

```bash
infraaudit notif update email --enabled true
infraaudit notif update slack --enabled false
```

| Flag | Description |
|------|-------------|
| `--enabled` | Enable or disable: `true`/`false` |

#### `notification history`

Show notification history.

```bash
infraaudit notif history
```

#### `notification send`

Send a test notification.

```bash
infraaudit notif send --channel email --message "Test notification"
```

| Flag | Description |
|------|-------------|
| `--channel` | Notification channel |
| `--message` | Notification message |

---

### webhook

Manage webhook integrations.

#### `webhook list`

List configured webhooks.

```bash
infraaudit webhook list
```

#### `webhook create`

Create a new webhook. Prompts interactively if flags are not provided.

```bash
infraaudit webhook create --name "Slack Alerts" --url https://hooks.slack.com/... --events drift.detected,alert.created
```

| Flag | Description |
|------|-------------|
| `--name` | Webhook name |
| `--url` | Webhook endpoint URL |
| `--secret` | Webhook signing secret |
| `--events` | Comma-separated event list |

#### `webhook get <id>`

Show webhook details.

```bash
infraaudit webhook get 1
```

#### `webhook delete <id>`

Delete a webhook.

```bash
infraaudit webhook delete 1
```

#### `webhook test <id>`

Send a test event to a webhook.

```bash
infraaudit webhook test 1
```

#### `webhook events`

List all available webhook event types.

```bash
infraaudit webhook events
```

---

## Shell Completion

Generate shell completion scripts for tab-completion support.

### Bash

```bash
# Current session
source <(infraaudit completion bash)

# Permanent (add to ~/.bashrc)
infraaudit completion bash > /etc/bash_completion.d/infraaudit
```

### Zsh

```bash
# Current session
source <(infraaudit completion zsh)

# Permanent (add to ~/.zshrc)
infraaudit completion zsh > "${fpath[1]}/_infraaudit"
```

### Fish

```bash
infraaudit completion fish | source

# Permanent
infraaudit completion fish > ~/.config/fish/completions/infraaudit.fish
```

### PowerShell

```powershell
infraaudit completion powershell | Out-String | Invoke-Expression
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `INFRAAUDIT_SERVER_URL` | Server URL | `http://localhost:8080` |
| `INFRAAUDIT_OUTPUT` | Default output format | `table` |

Environment variables override config file values. CLI flags override both.

**Precedence:** CLI flags > environment variables > config file > defaults

---

## Examples

### End-to-End Security Audit Workflow

```bash
# Setup
infraaudit config init
infraaudit auth login

# Connect AWS
infraaudit provider connect aws
infraaudit provider sync 1

# Security scan
infraaudit drift detect
infraaudit vuln scan

# Review findings
infraaudit drift list --severity critical
infraaudit vuln top
infraaudit alert list --severity high

# Remediate
infraaudit remediation suggest-drift 1
infraaudit remediation approve 1
infraaudit remediation execute 1

# Verify
infraaudit drift list --status resolved
infraaudit status
```

### Cost Optimization Workflow

```bash
# Sync cost data
infraaudit cost sync

# Analyze
infraaudit cost overview
infraaudit cost trends --period 30d
infraaudit cost anomalies
infraaudit cost forecast --days 90

# Get recommendations
infraaudit rec generate
infraaudit rec list --type cost
infraaudit rec savings

# Apply recommendation
infraaudit rec apply 10
```

### Compliance Assessment

```bash
# Enable framework
infraaudit compliance frameworks
infraaudit compliance enable cis-aws

# Run assessment
infraaudit compliance assess --framework cis-aws

# Review results
infraaudit compliance overview
infraaudit compliance failing-controls

# Export report
infraaudit compliance assessments
infraaudit compliance export 1
```

### CI/CD Integration

```bash
#!/bin/bash
# ci-security-check.sh

export INFRAAUDIT_SERVER_URL=https://api.infraaudit.dev

# Login with pre-stored credentials
infraaudit auth login --email "$CI_EMAIL" --password "$CI_PASSWORD"

# Run scans
infraaudit drift detect
infraaudit vuln scan

# Check for critical findings (JSON output for parsing)
CRITICAL=$(infraaudit drift list --severity critical -o json | jq 'length')
if [ "$CRITICAL" -gt 0 ]; then
  echo "FAIL: $CRITICAL critical drifts detected"
  infraaudit drift list --severity critical
  exit 1
fi

echo "PASS: No critical drifts"
```

### Scripting with JSON Output

```bash
# Get all resource IDs as a list
infraaudit resource list -o json | jq '.[].id'

# Count resources by type
infraaudit resource list -o json | jq 'group_by(.resource_type) | map({type: .[0].resource_type, count: length})'

# Get total estimated savings
infraaudit rec list -o json | jq '[.[].estimated_savings] | add'

# Export drift details to CSV
infraaudit drift list -o json | jq -r '.[] | [.id, .drift_type, .severity, .status] | @csv'
```
