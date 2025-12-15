# CID Tracker API Documentation

## Overview

CID Tracker exposes several HTTP endpoints for monitoring and management.

## Endpoints

### Health Check

**GET** `/health`

Returns basic health status of the CID Tracker service.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

**Status Codes:**
- `200` - Service is healthy
- `503` - Service is unhealthy

### Detailed Status

**GET** `/status`

Returns detailed status information including processing statistics.

**Response:**
```json
{
  "status": "running",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "statistics": {
    "logs_processed": 15420,
    "cids_extracted": 8934,
    "errors": 12,
    "last_processed": "2024-01-15T10:29:55Z"
  },
  "configuration": {
    "log_directory": "/var/log/app",
    "output_format": "json",
    "buffer_size": 1000,
    "poll_interval": "100ms"
  },
  "monitored_files": [
    {
      "path": "/var/log/app/application.log",
      "size": 104857600,
      "last_modified": "2024-01-15T10:29:50Z",
      "lines_processed": 8934
    }
  ]
}
```

### Metrics (Prometheus)

**GET** `/metrics`

Returns Prometheus-compatible metrics.

**Response:**
```
# HELP cidtracker_logs_processed_total Total number of log lines processed
# TYPE cidtracker_logs_processed_total counter
cidtracker_logs_processed_total 15420

# HELP cidtracker_cids_extracted_total Total number of CIDs extracted
# TYPE cidtracker_cids_extracted_total counter
cidtracker_cids_extracted_total 8934

# HELP cidtracker_errors_total Total number of processing errors
# TYPE cidtracker_errors_total counter
cidtracker_errors_total 12

# HELP cidtracker_file_size_bytes Current size of monitored log files
# TYPE cidtracker_file_size_bytes gauge
cidtracker_file_size_bytes{file="/var/log/app/application.log"} 104857600

# HELP cidtracker_processing_duration_seconds Time spent processing logs
# TYPE cidtracker_processing_duration_seconds histogram
cidtracker_processing_duration_seconds_bucket{le="0.001"} 5430
cidtracker_processing_duration_seconds_bucket{le="0.01"} 8920
cidtracker_processing_duration_seconds_bucket{le="0.1"} 8934
cidtracker_processing_duration_seconds_bucket{le="+Inf"} 8934
cidtracker_processing_duration_seconds_sum 8.934
cidtracker_processing_duration_seconds_count 8934
```

### Configuration

**GET** `/config`

Returns current configuration (non-sensitive values only).

**Response:**
```json
{
  "log_directory": "/var/log/app",
  "output_format": "json",
  "buffer_size": 1000,
  "poll_interval": "100ms",
  "cid_pattern": "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-5[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}",
  "output_destination": "stdout"
}
```

## Output Format

### JSON Output

When configured for JSON output, CID Tracker emits structured log entries:

```json
{
  "timestamp": "2024-01-15T10:30:00.123Z",
  "source_file": "/var/log/app/application.log",
  "line_number": 1234,
  "original_log": "2024-01-15 10:30:00 INFO [req-12345678-1234-5abc-9def-123456789012] Processing user request",
  "extracted_cid": "12345678-1234-5abc-9def-123456789012",
  "cid_type": "uuid_v5",
  "log_timestamp": "2024-01-15T10:30:00Z",
  "correlation_id": "cid_12345678_1234_5abc_9def_123456789012"
}
```

### Structured Output

When configured for structured output:

```
[2024-01-15T10:30:00.123Z] CID=12345678-1234-5abc-9def-123456789012 FILE=/var/log/app/application.log LINE=1234 TYPE=uuid_v5
```

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "error": "error description",
  "timestamp": "2024-01-15T10:30:00Z",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**
- `INVALID_REQUEST` - Malformed request
- `NOT_FOUND` - Requested resource not found
- `INTERNAL_ERROR` - Internal server error
- `SERVICE_UNAVAILABLE` - Service temporarily unavailable