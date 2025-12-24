<p align="center">
  <img src="docs/assets/logo.svg" alt="CID Tracker Logo" width="120" />
</p>

<h1 align="center">CID Tracker</h1>

<p align="center">
  <strong>Extract correlation IDs from your logs. Automatically.</strong>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="#project-status"><img src="https://img.shields.io/badge/Status-Alpha-orange.svg" alt="Status"></a>
</p>

<p align="center">
  <a href="#what-is-cid-tracker">What is it?</a> •
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#how-it-works">How it Works</a> •
  <a href="#contributing">Contributing</a>
</p>

* * *

## What is CID Tracker?

> *"Our microservices generate thousands of log lines per minute. When something breaks, I spend more time correlating logs across services than actually fixing the problem."*

Sound familiar? **CID Tracker** solves this by automatically extracting correlation IDs (CIDs) from your application logs and outputting them as structured data — ready for your observability pipeline.

**Deploy it as a sidecar container.** It watches your log files, extracts UUIDs used as correlation IDs, validates them, and streams structured output to stdout. No code changes required in your application.

### Before CID Tracker

```
# Scattered across multiple log files, multiple services...
2024-01-15 10:30:00 INFO [auth] CID:550e8400-e29b-51d4-a716-446655440000 User login
2024-01-15 10:30:01 INFO [orders] CID:550e8400-e29b-51d4-a716-446655440000 Fetching cart
2024-01-15 10:30:02 ERROR [payments] CID:550e8400-e29b-51d4-a716-446655440000 Payment failed
# Good luck finding these manually...
```

### After CID Tracker

```json
{"cid":"550e8400-e29b-51d4-a716-446655440000","log_file":"auth.log","timestamp":"2024-01-15T10:30:00Z",...}
{"cid":"550e8400-e29b-51d4-a716-446655440000","log_file":"orders.log","timestamp":"2024-01-15T10:30:01Z",...}
{"cid":"550e8400-e29b-51d4-a716-446655440000","log_file":"payments.log","timestamp":"2024-01-15T10:30:02Z",...}
```

Pipe this to Elasticsearch, Loki, or any log aggregator — now you can filter by CID instantly.

* * *

## Features

<table>
<tr>
<td width="50%" valign="top">

### Core
- **Real-time monitoring** — Uses filesystem events, not polling
- **Pattern matching** — Configurable regex for your log format
- **UUID validation** — Validates extracted IDs, supports v5 enforcement
- **Multiple outputs** — JSON or structured text

</td>
<td width="50%" valign="top">

### Operations
- **Sidecar-ready** — Minimal footprint, container-native
- **Graceful shutdown** — Clean handling of SIGTERM
- **Prometheus metrics** — Built-in `/metrics` endpoint
- **Zero dependencies** — Single binary, no runtime requirements

</td>
</tr>
</table>

* * *

## Quick Start

### See It In Action (30 seconds)

```bash
git clone https://github.com/psenger/cidtracker.git
cd cidtracker/examples
docker compose up --build
```

Watch the output:

```
sample-app     | 2024-01-15 10:30:00 INFO [auth-service] CID:550e8400-e29b-51d4-a716-446655440000 Processing user request
cidtracker     | {"cid":"550e8400-e29b-51d4-a716-446655440000","uuid":"550e8400-e29b-51d4-a716-446655440000","timestamp":"2024-01-15T10:30:00Z","log_file":"application.log",...}
```

The sample app generates logs with CIDs. CID Tracker extracts them in real-time.

Press `Ctrl+C` to stop, then `docker compose down -v` to clean up.

### Install

**From Source:**
```bash
go install github.com/psenger/cidtracker@latest
```

**Build Locally:**
```bash
git clone https://github.com/psenger/cidtracker.git
cd cidtracker
go build -o cidtracker .
```

**Docker:**
```bash
docker build -t cidtracker:latest .
```

### Basic Usage

```bash
# Monitor a directory, output JSON
./cidtracker -log-path=/var/log/app -output=json

# With verbose logging
./cidtracker -log-path=/var/log/app -output=json -verbose
```

* * *

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        Your Application                         │
│                                                                 │
│   logger.info("CID:{} Processing order", correlationId)         │
│                              │                                  │
│                              ▼                                  │
│                    /var/log/app/app.log                         │
└─────────────────────────────────────────────────────────────────┘
                               │
                               │ (shared volume)
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                         CID Tracker                             │
│                                                                 │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│   │  LogMonitor  │───▶│  Extractor   │───▶│  Validator   │      │
│   │  (fsnotify)  │    │  (regex)     │    │  (UUID v5)   │      │
│   └──────────────┘    └──────────────┘    └──────────────┘      │
│                                                  │              │
│                                                  ▼              │
│                                           ┌──────────────┐      │
│                                           │    stdout    │      │
│                                           │    (JSON)    │      │
│                                           └──────────────┘      │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                    Log Aggregator / SIEM
                (Elasticsearch, Loki, Splunk, etc.)
```

1. **LogMonitor** watches for new `.log` files and tails existing ones
2. **Extractor** applies regex patterns to find `CID:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
3. **Validator** confirms it's a valid UUID (optionally version 5 only)
4. **Output** streams structured JSON to stdout for your pipeline

* * *

## Use Cases

> *"I'm a **platform engineer** building a Kubernetes logging stack. I need to enrich logs with correlation metadata before they hit Elasticsearch."*

Deploy CID Tracker as a sidecar. Mount the log volume read-only. Pipe stdout to Fluentd/Fluent Bit.

> *"I'm a **developer** debugging a distributed system. I need to trace a request across 5 microservices."*

Add CID Tracker to your docker-compose. Filter aggregated logs by the extracted CID.

> *"I'm an **SRE** setting up alerting. I need to detect when the same correlation ID appears in error logs across multiple services."*

Feed CID Tracker output to your SIEM. Create correlation rules based on CID frequency in error streams.

* * *

## Deployment

### Docker Compose (Sidecar)

```yaml
version: '3.8'
services:
  your-app:
    image: your-app:latest
    volumes:
      - logs:/var/log/app

  cidtracker:
    image: cidtracker:latest
    volumes:
      - logs:/var/log/app:ro
    command: ["-log-path=/var/log/app", "-output=json"]

volumes:
  logs:
```

### Kubernetes

```yaml
spec:
  containers:
  - name: app
    image: your-app:latest
    volumeMounts:
    - name: logs
      mountPath: /var/log/app

  - name: cidtracker
    image: cidtracker:latest
    args: ["-log-path=/var/log/app", "-output=json"]
    volumeMounts:
    - name: logs
      mountPath: /var/log/app
      readOnly: true
    resources:
      limits:
        memory: "64Mi"
        cpu: "50m"

  volumes:
  - name: logs
    emptyDir: {}
```

See [Deployment Guide](docs/deployment.md) for complete examples.

* * *

## Configuration

| Flag        | Environment Variable         | Default        | Description                           |
|-------------|------------------------------|----------------|---------------------------------------|
| `-log-path` | `CIDTRACKER_LOG_DIR`         | `/var/log/app` | Directory to monitor                  |
| `-output`   | `CIDTRACKER_OUTPUT_FORMAT`   | `json`         | Output format (`json` / `structured`) |
| `-verbose`  | `CIDTRACKER_LOG_LEVEL=debug` | `false`        | Enable debug logging                  |

* * *

## Project Status

> **Alpha** — Core functionality works. Not yet recommended for production.

### What Works
- [x] File monitoring with fsnotify
- [x] CID extraction via regex
- [x] UUID validation (all versions)
- [x] JSON and structured output
- [x] Graceful shutdown
- [x] Docker deployment

### In Progress
- [ ] HTTP server with `/health` and `/metrics`
- [ ] Configurable CID patterns via file
- [ ] Log rotation handling

### Planned
- [ ] Multi-file correlation
- [ ] Custom output destinations
- [ ] UUID version 7 support

* * *

## Documentation

| Document                               | Description                           |
|----------------------------------------|---------------------------------------|
| [API Reference](docs/api.md)           | HTTP endpoints (when enabled)         |
| [Deployment Guide](docs/deployment.md) | Docker, Kubernetes, and configuration |
| [Examples](examples/)                  | Working docker-compose demo           |
| [Contributing](CONTRIBUTING.md)        | How to contribute                     |

* * *

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**The short version:**

1. **Open an issue first** — Let's discuss before you code
2. Fork → Branch → Code → Test
3. Ensure **75% test coverage** minimum
4. Submit PR referencing the issue

```bash
# Run tests
go test ./...

# Check coverage
go test -cover ./...
```

* * *

## License

MIT License — see [LICENSE](LICENSE) for details.

* * *

## Acknowledgments

Built with:
- [fsnotify](https://github.com/fsnotify/fsnotify) — File system notifications
- [google/uuid](https://github.com/google/uuid) — UUID parsing
- [logrus](https://github.com/sirupsen/logrus) — Structured logging

* * *

<p align="center">
  <sub>Built with Go. Made for observability.</sub>
</p>
