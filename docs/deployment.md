# CID Tracker Deployment Guide

## Docker Deployment

### Building the Image

```bash
docker build -t cidtracker:latest .
```

### Running as Sidecar

Deploy CID Tracker as a sidecar container alongside your main application:

```yaml
version: '3.8'
services:
  app:
    image: your-app:latest
    volumes:
      - app-logs:/var/log/app
  
  cidtracker:
    image: cidtracker:latest
    volumes:
      - app-logs:/var/log/app:ro
      - ./config.yaml:/etc/cidtracker/config.yaml
    environment:
      - CIDTRACKER_LOG_DIR=/var/log/app
      - CIDTRACKER_OUTPUT_FORMAT=json
    depends_on:
      - app

volumes:
  app-logs:
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-with-cidtracker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: your-app:latest
        volumeMounts:
        - name: logs
          mountPath: /var/log/app
      
      - name: cidtracker
        image: cidtracker:latest
        volumeMounts:
        - name: logs
          mountPath: /var/log/app
          readOnly: true
        - name: config
          mountPath: /etc/cidtracker
        env:
        - name: CIDTRACKER_LOG_DIR
          value: "/var/log/app"
        - name: CIDTRACKER_OUTPUT_FORMAT
          value: "json"
        resources:
          limits:
            memory: "128Mi"
            cpu: "100m"
          requests:
            memory: "64Mi"
            cpu: "50m"
      
      volumes:
      - name: logs
        emptyDir: {}
      - name: config
        configMap:
          name: cidtracker-config
```

## Configuration

### Environment Variables

| Variable                   | Description                     | Default              |
|----------------------------|---------------------------------|----------------------|
| `CIDTRACKER_LOG_DIR`       | Directory to monitor for logs   | `/var/log`           |
| `CIDTRACKER_OUTPUT_FORMAT` | Output format (json/structured) | `json`               |
| `CIDTRACKER_BUFFER_SIZE`   | Log processing buffer size      | `1000`               |
| `CIDTRACKER_POLL_INTERVAL` | File polling interval           | `100ms`              |
| `CIDTRACKER_CID_PATTERN`   | Custom CID regex pattern        | (default U5 pattern) |

### Health Checks

CID Tracker exposes health endpoints:

- `/health` - Basic health check
- `/metrics` - Prometheus metrics
- `/status` - Detailed status information

## Monitoring Integration

### Prometheus Metrics

```yaml
scrape_configs:
  - job_name: 'cidtracker'
    static_configs:
      - targets: ['cidtracker:8080']
    metrics_path: /metrics
```

### Log Aggregation

Configure your log aggregation system to consume CID Tracker output:

```yaml
# Fluentd configuration
<source>
  @type tail
  path /var/log/cidtracker/output.json
  pos_file /var/log/fluentd/cidtracker.log.pos
  tag cidtracker
  format json
</source>

<match cidtracker>
  @type elasticsearch
  host elasticsearch.logging.svc.cluster.local
  port 9200
  index_name cidtracker
</match>
```

## Troubleshooting

### Common Issues

1. **No logs being processed**
   - Verify volume mounts are correct
   - Check file permissions (CID Tracker needs read access)
   - Ensure log directory exists

2. **High memory usage**
   - Reduce `CIDTRACKER_BUFFER_SIZE`
   - Increase `CIDTRACKER_POLL_INTERVAL`
   - Check for log rotation issues

3. **Missing CID extractions**
   - Verify CID pattern matches your log format
   - Check UUID validation settings
   - Review sample logs for pattern alignment

### Debug Mode

Enable debug logging:

```bash
docker run -e CIDTRACKER_LOG_LEVEL=debug cidtracker:latest
```