# Contributing to CID Tracker

Thank you for your interest in contributing to CID Tracker! This document provides guidelines and instructions for contributing.

## Before You Start

**Please open an issue first** before starting work on any significant changes. This helps us:

- Discuss the proposed changes and ensure they align with the project's direction
- Avoid duplicate effort if someone else is already working on it
- Provide guidance on implementation approach

For bug fixes, open a [Bug Report](https://github.com/psenger/cidtracker/issues/new?template=bug_report.md).
For new features, open a [Feature Request](https://github.com/psenger/cidtracker/issues/new?template=feature_request.md).

## Try It First

Before diving into the code, run the example to understand what CID Tracker does:

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/psenger/cidtracker.git
cd cidtracker/examples

# Start the demo
docker compose up --build
```

You'll see two containers running:

1. **sample-app** — Generates log lines with correlation IDs:
   ```
   2024-01-15 10:30:00 INFO [auth-service] CID:550e8400-e29b-51d4-a716-446655440000 request_id=42356 Processing user request
   ```

2. **cidtracker** — Extracts and outputs structured JSON:
   ```json
   {"cid":"550e8400-e29b-51d4-a716-446655440000","uuid":"550e8400-e29b-51d4-a716-446655440000","timestamp":"2024-01-15T10:30:00Z","log_file":"application.log",...}
   ```

Press `Ctrl+C` to stop, then `docker compose down -v` to clean up.

### Running Locally (Without Docker)

**Terminal 1 — Build and run CID Tracker:**
```bash
cd cidtracker
go build -o cidtracker .
mkdir -p /tmp/demo-logs
./cidtracker -log-path=/tmp/demo-logs -output=json -verbose
```

**Terminal 2 — Generate sample logs:**
```bash
while true; do
  CID=$(uuidgen | tr '[:upper:]' '[:lower:]')
  echo "$(date '+%Y-%m-%d %H:%M:%S') INFO [demo] CID:$CID Processing request" >> /tmp/demo-logs/app.log
  sleep 2
done
```

Watch Terminal 1 — you'll see CID Tracker extract each correlation ID as logs are written.

* * *

## Getting Started (Development)

### Prerequisites

- Go 1.21 or later
- Git
- Docker (optional, for running examples)

### Setup

1. **Fork the repository** on GitHub

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/cidtracker.git
   cd cidtracker
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/psenger/cidtracker.git
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Verify the build**
   ```bash
   go build ./...
   ```

6. **Run tests**
   ```bash
   go test ./...
   ```

7. **Run the example** (to verify everything works)
   ```bash
   cd examples
   docker compose up --build
   ```

## Development Workflow

### Creating a Branch

Create a branch from `main` for your work:

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Making Changes

1. Write your code following the existing style and patterns
2. Add or update tests as needed
3. Ensure all tests pass
4. Update documentation if applicable

### Code Quality Requirements

#### Test Coverage

We maintain a **minimum test coverage threshold of 75%**. Your changes must not decrease the overall coverage below this threshold.

Check coverage before submitting:

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# View coverage in browser (optional)
go tool cover -html=coverage.out
```

#### Code Standards

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common issues
- Keep functions focused and reasonably sized
- Add comments for exported functions and complex logic

```bash
# Format code
go fmt ./...

# Run vet
go vet ./...
```

### Commit Messages

Write clear, concise commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Fix bug" not "Fixes bug")
- Keep the first line under 72 characters
- Reference issues when applicable

Examples:
```
Add UUID version 7 support

Implement extraction and validation for UUID v7 format.
Closes #42
```

```
Fix file handle leak in log monitor

File handles were not being closed when log files were deleted.
Fixes #37
```

## Submitting Changes

### Pull Request Process

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Open a Pull Request** on GitHub with:
   - A clear title describing the change
   - A description explaining what and why
   - Reference to the related issue (e.g., "Closes #123")

4. **Respond to review feedback** and make updates as needed

### Pull Request Checklist

Before submitting, ensure:

- [ ] All tests pass (`go test ./...`)
- [ ] Coverage remains above 75% (`go test -cover ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No vet issues (`go vet ./...`)
- [ ] Documentation is updated (if applicable)
- [ ] Commit messages are clear and descriptive
- [ ] PR references the related issue

## Project Structure

```
cidtracker/
├── main.go              # Application entry point
├── tracker.go           # Main CIDTracker implementation
├── pkg/
│   ├── config/          # Configuration management
│   ├── extractor/       # CID extraction logic
│   ├── models/          # Data structures
│   ├── monitor/         # File monitoring
│   ├── processor/       # Processing pipeline
│   └── validator/       # UUID validation
├── docs/                # Documentation
└── .github/             # GitHub templates and workflows
```

## Getting Help

- **Questions:** Open a [Discussion](https://github.com/psenger/cidtracker/discussions)
- **Bugs:** Open a [Bug Report](https://github.com/psenger/cidtracker/issues/new?template=bug_report.md)
- **Features:** Open a [Feature Request](https://github.com/psenger/cidtracker/issues/new?template=feature_request.md)

## License

By contributing to CID Tracker, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing!
