# Klaudia Sync Go CLI and Docker Action

[![GitHub Action](https://img.shields.io/badge/GitHub-Action-blue.svg)](https://github.com/komodorio/custom-komodor-integrations/tree/master/klaudia-sync-action)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A production-grade Go sync tool for Komodor's Klaudia API with **full CRUD operations**, detailed logging, and CI/CD integration. The same code runs as a Docker action and as a local CLI.

## ✨ Key Features

- **🔄 Full CRUD Operations**
  - Upload new files
  - Update changed files
  - Delete removed files
  - List and compare remote files

- **📝 Structured Logging**
  - Timestamped operation logs suitable for CI/CD
  - Multiple log levels (debug, info, warn, error)
  - Detailed operation summary
  - Full audit trail of changes

- **⚙️ Flexible Configuration**
  - Support for `knowledge-base` and `blueprints` file types
  - File extension filtering (sync only `.md`, `.txt`, etc.)
  - Optional recursive directory traversal
  - Configurable API endpoint
  - Works from GitHub Actions, Docker, GitLab, Jenkins, or a local shell

- **🧪 Dry Run Mode**
  - Preview all changes without applying them
  - Safe testing before production

- **📊 Detailed Outputs**
  - Number of files uploaded, updated, deleted
  - Human-readable sync summary
  - Complete operation log for debugging

## 📋 Quick Start

### 1. Set Your API Key

```yaml
# Add to GitHub Secrets (Settings → Secrets and variables → Actions)
KOMODOR_API_KEY: your-api-key-here
```

### 2. Create Workflow File

```yaml
name: Sync KB to Komodor

on:
  push:
    paths: ['kb/**']
    branches: [main]

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Sync Knowledge Base
        uses: komodorio/klaudia-sync-action@v1
        with:
          directory: ./kb
          file-type: knowledge-base
          api-key: ${{ secrets.KOMODOR_API_KEY }}
          file-extensions: '.md'
```

### 3. Run Locally

```bash
go run ./cmd/klaudia-sync --directory ./kb --file-type knowledge-base --api-key "$KOMODOR_API_KEY"
```

Or with Docker:

```bash
docker build -t klaudia-sync-action .
docker run --rm -e KOMODOR_API_KEY="$KOMODOR_API_KEY" -e KLAUDIA_DIRECTORY=/workspace/kb -e KLAUDIA_FILE_TYPE=knowledge-base -v "$PWD:/workspace" klaudia-sync-action
```

### 4. Push Your Changes

```bash
git add kb/
git commit -m "Add KB article"
git push origin main
```

## 📖 Documentation

### For Action Users

See the **[klaudia-sync-action README](klaudia-sync-action/README.md)** for:

- Complete input/output reference
- Advanced configuration options
- Troubleshooting guide
- Example workflows

### For Example Setup

See **[KlaudiaKBSync/README.md](KlaudiaKBSync/README.md)** for:

- Step-by-step setup instructions
- Pre-built workflow examples
- Best practices
- Customisation guide

## 🏗️ Architecture

```plain
klaudia-sync-action/
├── action.yml              # Docker action definition
├── cmd/klaudia-sync/       # CLI entrypoint
├── internal/klaudia/       # Core sync engine and tests
├── Dockerfile              # Container runtime
└── README.md               # Action documentation

KlaudiaKBSync/             # Example setup
├── kb/                     # Sample KB articles
├── .github/
│   └── workflows/
│       └── sync-kb.yml    # Pre-configured workflow
└── README.md               # Setup guide
```

## 🔑 Authentication

Store your Komodor API key as a GitHub Secret:

1. Go to **Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Name: `KOMODOR_API_KEY`
4. Paste your API key
5. Use in workflows: `${{ secrets.KOMODOR_API_KEY }}`

## 📊 Example Output

```plain
[2024-01-15T10:30:45.123Z] [INFO] Starting Klaudia Sync Action
[2024-01-15T10:30:45.456Z] [INFO] ================================================================================
[2024-01-15T10:30:45.457Z] [INFO] VALIDATING INPUTS
[2024-01-15T10:30:45.457Z] [INFO] ✓ API key provided
[2024-01-15T10:30:45.458Z] [INFO] ✓ File type is valid: knowledge-base
[2024-01-15T10:30:45.459Z] [INFO] ✓ Directory exists: ./kb
[2024-01-15T10:30:46.000Z] [INFO] ================================================================================
[2024-01-15T10:30:46.001Z] [INFO] LISTING REMOTE FILES
[2024-01-15T10:30:46.234Z] [INFO] Found 3 remote files
[2024-01-15T10:30:46.235Z] [INFO] ================================================================================
[2024-01-15T10:30:46.236Z] [INFO] PERFORMING SYNC OPERATION
[2024-01-15T10:30:46.456Z] [INFO] ✓ Uploaded: new-guide.md
[2024-01-15T10:30:46.789Z] [INFO] ✓ Updated: existing-guide.md
[2024-01-15T10:30:47.012Z] [INFO] ✓ Deleted: deprecated-guide.md
[2024-01-15T10:30:47.234Z] [INFO] ================================================================================
[2024-01-15T10:30:47.235Z] [INFO] SYNC SUMMARY
[2024-01-15T10:30:47.235Z] [INFO] Files uploaded: 1
[2024-01-15T10:30:47.236Z] [INFO] Files updated: 1
[2024-01-15T10:30:47.237Z] [INFO] Files deleted: 1
[2024-01-15T10:30:47.238Z] [INFO] Total changes: 3
```

## 🛠️ Use Cases

### Knowledge Base Management

Automatically sync your team's knowledge base to Komodor:

```yaml
directory: ./docs/kb
file-type: knowledge-base
file-extensions: '.md'
```

### Blueprint Configuration

Sync deployment blueprints:

```yaml
directory: ./blueprints
file-type: blueprints
file-extensions: '.yaml,.yml'
```

### Daily Synchronisation

Keep files in sync with scheduled workflows:

```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
```

### Pre-deployment Review

Dry-run before production:

```yaml
dry-run: 'true'  # Preview changes first
```

## ⚡ Features in Detail

### Bidirectional Sync

- **Local → Remote**: Upload new/updated files
- **Remote → Local comparison**: Delete files no longer in local directory
- **Smart updates**: Only modifies changed files

### File Filtering

- Include specific extensions: `.md`, `.txt`, `.yaml`
- Sync all files or just matching patterns
- Recursive directory traversal (configurable)

### Error Handling

- Detailed error messages for each failed operation
- Graceful degradation (continues on individual failures)
- Separate failure counter in summary

### Operation Logging

- Timestamped logs for audit trails
- Multiple verbosity levels
- Structured output for CI/CD systems
- Complete operation history in outputs

## 🔒 Security

- API key passed securely via GitHub Secrets
- HTTPS-only API communication
- No credentials logged or exposed
- Action runs with minimal permissions

## 📝 Inputs Reference

| Input | Required | Default | Description |
| ------- | ---------- | --------- | ------------- |
| `directory` | ✅ | - | Local directory to sync |
| `file-type` | ✅ | - | `knowledge-base` or `blueprints` |
| `api-key` | ✅ | - | Komodor API key |
| `api-base-url` | ❌ | `https://api.komodor.com` | API endpoint |
| `recursive` | ❌ | `true` | Traverse subdirectories |
| `dry-run` | ❌ | `false` | Preview without applying |
| `file-extensions` | ❌ | `` | Comma-separated filter (e.g., `.md,.txt`) |

## 📤 Outputs Reference

| Output | Description |
| -------- | ------------- |
| `summary` | Human-readable sync summary |
| `files-uploaded` | Count of uploaded files |
| `files-updated` | Count of updated files |
| `files-deleted` | Count of deleted files |
| `operation-log` | Full timestamped operation log |

## 🚀 Advanced Examples

### Sync Multiple File Types

```yaml
jobs:
  sync-kb:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: komodorio/klaudia-sync-action@v1
        with:
          directory: ./kb
          file-type: knowledge-base
          api-key: ${{ secrets.KOMODOR_API_KEY }}

  sync-blueprints:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: komodorio/klaudia-sync-action@v1
        with:
          directory: ./blueprints
          file-type: blueprints
          api-key: ${{ secrets.KOMODOR_API_KEY }}
```

### Slack Notification on Sync

```yaml
- name: Sync to Komodor
  id: sync
  uses: komodorio/klaudia-sync-action@v1
  with:
    directory: ./kb
    file-type: knowledge-base
    api-key: ${{ secrets.KOMODOR_API_KEY }}

- name: Notify Slack
  if: always()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "Knowledge Base Sync Complete",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "${{ steps.sync.outputs.summary }}"
            }
          }
        ]
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Pull Request Comment with Results

```yaml
- name: Comment PR
  if: github.event_name == 'pull_request'
  uses: actions/github-script@v6
  with:
    script: |
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: '${{ steps.sync.outputs.operation-log }}'
      })
```

## 🐛 Troubleshooting

### API Key Issues

```plain
Error: API request failed: 401 Unauthorized
```

- Verify API key in GitHub Secrets
- Check key hasn't expired in Komodor
- Ensure correct secret name: `KOMODOR_API_KEY`

### Directory Not Found

```plain
Error: Directory does not exist: ./kb
```

- Check path is relative to repository root
- Ensure directory exists in checked-out code
- Verify path in `directory` input

### No Files Synced

```plain
[INFO] Found 0 files to sync
```

- Check file extensions filter
- Verify files exist in directory
- Enable recursive mode if using subdirectories

See [full troubleshooting guide](klaudia-sync-action/README.md#troubleshooting).

## 📚 Learn More

- [Klaudia Sync Action README](klaudia-sync-action/README.md) - Complete action documentation
- [KlaudiaKBSync Setup Guide](KlaudiaKBSync/README.md) - Example configuration and workflows
- [Komodor API Documentation](https://api.komodor.com/api/docs/)

## 🤝 Contributing

Found a bug or have a feature request? Open an issue in the repository.

## 📄 License

Apache License 2.0 - See LICENSE file for details

---

**Built by Komodor** | [Visit Komodor](https://komodor.com)
