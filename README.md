# CID Tracker

Docker sidecar for observability and correlation of CID-tagged logs with UUID extraction. Monitors mounted docker logs to extract correlation IDs (CID:xxxxx) and timestamps for distributed tracing and request correlation across microservices.

## Features

- Real-time log monitoring via volume mounts
- CID pattern extraction with UUID validation
- Timestamp correlation and indexing
- Structured JSON output for downstream observability tools
- Graceful shutdown handling
- Configurable output formats

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd cidtracker

# Install dependencies
go mod download

# Build the application
go build -o cidtracker
```

## Usage

### Basic Usage

```bash
# Monitor logs in /var/log/app directory
./cidtracker -log-path=/var/log/app

# Enable verbose logging with structured output
./cidtracker -log-path=/var/log/app -verbose -output=structured
```

### Docker Sidecar

```bash
# Run as sidecar container with mounted log volume
docker run -v /path/to/logs:/var/log/app cidtracker
```

### Command Line Options

- `-log-path`: Path to mounted docker logs directory (default: `/var/log/app`)
- `-output`: Output format - `json` or `structured` (default: `json`)
- `-verbose`: Enable verbose logging

### Expected Log Format

The tracker looks for correlation IDs in the format:
```
CID:550e8400-e29b-41d4-a716-446655440000
```

### Output Format

**JSON Output:**
```json
{
  "cid": "550e8400-e29b-41d4-a716-446655440000",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2023-12-07T10:30:45Z",
  "log_file": "app.log",
  "raw_message": "Processing request CID:550e8400-e29b-41d4-a716-446655440000",
  "processed_at": "2023-12-07T10:30:45.123Z"
}
```

**Structured Output:**
```
[2023-12-07T10:30:45Z] CID:550e8400-e29b-41d4-a716-446655440000 FILE:app.log
```

## Integration

Pipe output to your observability stack:
```bash
./cidtracker -log-path=/var/log/app | your-log-processor
```