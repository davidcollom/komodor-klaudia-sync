# Klaudia Sync

`klaudia-sync` is a Go CLI and Docker-based GitHub Action for synchronising local documentation into Komodor Klaudia with full CRUD behaviour.

It supports two remote file types:

- `knowledge-base`
- `blueprint`

## What It Does

- uploads new files
- updates changed files
- deletes remote files that no longer exist locally
- filters by extension when required
- supports dry-run mode
- emits structured logs via `logrus`
- retries transient HTTP failures with `retryablehttp`

## Supported Formats

- `.md`, `.markdown`
- `.pdf`
- `.txt`
- `.doc`, `.docx`
- `.csv`
- `.json`
- `.yaml`, `.yml`

Maximum file size is `54,945,382` bytes per file.

## Local CLI

Run directly:

```bash
go run ./cmd/klaudia-sync sync \
  --directory ./example/kb \
  --file-type knowledge-base \
  --api-key "$KOMODOR_API_KEY"
```

Preview a blueprint sync with debug logs:

```bash
go run ./cmd/klaudia-sync sync \
  --directory ./example/blueprints \
  --file-type blueprint \
  --api-key "$KOMODOR_API_KEY" \
  --dry-run \
  --debug
```

Useful flags:

- `--recursive`
- `--dry-run`
- `--debug`
- `--file-extensions`
- `--api-base-url`

The CLI also reads these environment variables:

- `KOMODOR_API_KEY`
- `KOMODOR_API_BASE_URL`
- `KLAUDIA_DIRECTORY`
- `KLAUDIA_FILE_TYPE`
- `KLAUDIA_RECURSIVE`
- `KLAUDIA_DRY_RUN`
- `KLAUDIA_DEBUG`
- `KLAUDIA_FILE_EXTENSIONS`

## Docker

Build and run locally:

```bash
docker build -t klaudia-sync-action .

docker run --rm \
  -e KOMODOR_API_KEY="$KOMODOR_API_KEY" \
  -e KLAUDIA_DIRECTORY=/workspace/example/kb \
  -e KLAUDIA_FILE_TYPE=knowledge-base \
  -v "$PWD:/workspace" \
  klaudia-sync-action
```

## GitHub Action

The action is published from this repository and currently points to the prebuilt GHCR image declared in [action.yml](action.yml).

Example usage:

```yaml
name: Sync Knowledge Base

on:
  push:
    branches: [main]
    paths:
      - 'example/kb/**'

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Sync to Klaudia
        uses: davidcollom/komodor-klaudia-sync@v1
        with:
          directory: ./example/kb
          file-type: knowledge-base
          api-key: ${{ secrets.KOMODOR_API_KEY }}
          dry-run: 'false'
```

Blueprint example:

```yaml
- name: Preview blueprint sync
  uses: davidcollom/komodor-klaudia-sync@v1
  with:
    directory: ./example/blueprints
    file-type: blueprint
    api-key: ${{ secrets.KOMODOR_API_KEY }}
    dry-run: 'true'
    debug: 'true'
```

## Inputs

| Input | Description | Required | Default |
| ------- | ------------- | ---------- | --------- |
| `directory` | Local directory path to sync | Yes | - |
| `file-type` | `knowledge-base` or `blueprint` | Yes | `knowledge-base` |
| `api-key` | Komodor API key | Yes | - |
| `api-base-url` | Komodor API base URL | No | `https://api.komodor.com` |
| `recursive` | Recurse into subdirectories | No | `true` |
| `dry-run` | Preview changes without applying them | No | `false` |
| `debug` | Enable debug logging | No | `false` |
| `file-extensions` | Comma-separated extension filter | No | empty |

## Outputs

| Output | Description |
| -------- | ------------- |
| `summary` | Human-readable summary |
| `files-uploaded` | Number of files uploaded |
| `files-updated` | Number of files updated |
| `files-deleted` | Number of files deleted |
| `operation-log` | Detailed operation log |

## Logging And Error Handling

- info logs are shown by default
- debug file discovery is shown with `--debug`
- HTTP retries are applied for transient failures
- API errors include the HTTP method, request path, status, and Klaudia `request_id` when available

Example error:

```text
api GET /api/v2/klaudia/files/blueprint failed with 400 Bad Request (400): bad request: 400 Bad Request (request_id=...)
```

## Release Notes

Releases are built with GoReleaser and published to GitHub Releases and GHCR. The GitHub Action references the floating major image tag, which is updated by the release workflow before tagging.

## Repository Examples

- [example/kb](example/kb) contains knowledge-base runbooks
- [example/blueprints](example/blueprints) contains architecture and release-management blueprints

## License

Apache 2.0
