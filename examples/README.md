# CID Tracker Examples

This directory contains a complete working example demonstrating CID Tracker in action.

## What's Included

```
examples/
├── docker-compose.yml    # Orchestrates both containers
├── log-generator/        # Sample app that generates logs with CIDs
│   ├── main.go
│   └── Dockerfile
└── README.md             # This file
```

## Quick Start

### 1. Start the Example

From the `examples/` directory:

```bash
docker compose up --build
```

This starts two containers:
- **sample-app** — Generates log entries with correlation IDs every 3 seconds
- **cidtracker** — Monitors the logs and extracts/validates the CIDs

### 2. Watch It Work

You'll see output like this:

**Sample App (log-generator):**
```
2024-01-15 10:30:00.123 INFO [auth-service] CID:550e8400-e29b-51d4-a716-446655440000 request_id=42356 Processing user request
2024-01-15 10:30:03.456 DEBUG [order-service] CID:660e8400-e29b-51d4-b716-446655440001 request_id=78234 Fetching data from database
```

**CID Tracker (cidtracker):**
```json
{"cid":"550e8400-e29b-51d4-a716-446655440000","uuid":"550e8400-e29b-51d4-a716-446655440000","timestamp":"2024-01-15T10:30:00Z","log_file":"application.log","raw_message":"2024-01-15 10:30:00.123 INFO [auth-service] CID:550e8400-e29b-51d4-a716-446655440000 request_id=42356 Processing user request","processed_at":"2024-01-15T10:30:00.150Z"}
```

### 3. Stop the Example

```bash
docker compose down
```

To also remove the volumes:

```bash
docker compose down -v
```

## Understanding the Output

CID Tracker extracts each correlation ID and outputs structured JSON containing:

| Field | Description |
|-------|-------------|
| `cid` | The extracted correlation ID |
| `uuid` | The validated UUID value |
| `timestamp` | When the log was written |
| `log_file` | Source log file name |
| `raw_message` | The complete original log line |
| `processed_at` | When CID Tracker processed it |

## Customizing the Example

### Change Log Frequency

Edit `docker-compose.yml`:

```yaml
log-generator:
  environment:
    - LOG_INTERVAL=1s  # Generate logs every second
```

### Use Structured Output Instead of JSON

```yaml
cidtracker:
  command: ["-log-path=/var/log/app", "-output=structured", "-verbose"]
```

Output becomes:
```
[2024-01-15T10:30:00Z] CID:550e8400-e29b-51d4-a716-446655440000 FILE:application.log
```

### View the Raw Log Files

The logs are stored in a Docker volume. To inspect them:

```bash
# Find the volume
docker volume ls | grep app-logs

# Inspect the contents
docker run --rm -v examples_app-logs:/logs alpine cat /logs/application.log
```

## Running Without Docker

If you want to run locally without Docker:

### Terminal 1 — Run CID Tracker
```bash
cd /path/to/cidtracker
go build -o cidtracker .
mkdir -p /tmp/demo-logs
./cidtracker -log-path=/tmp/demo-logs -output=json -verbose
```

### Terminal 2 — Generate Sample Logs
```bash
# Simple log generator using bash
while true; do
  UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
  echo "$(date '+%Y-%m-%d %H:%M:%S') INFO [demo] CID:$UUID Processing request" >> /tmp/demo-logs/app.log
  sleep 2
done
```

## What This Demonstrates

1. **Sidecar Pattern** — CID Tracker runs alongside your app, sharing a log volume
2. **Real-time Processing** — Logs are processed as they're written
3. **UUID Validation** — Only valid UUIDs (specifically v5) are extracted
4. **Structured Output** — Ready for ingestion by log aggregators

## Next Steps

- Pipe CID Tracker output to your log aggregator (Fluentd, Logstash, etc.)
- Use the `/metrics` endpoint with Prometheus
- Add multiple log-generating services to see correlation across services
