# Go AWS Project

This is a Go project template with AWS SDK integration.

## Prerequisites

- Go 1.21 or later
- AWS credentials configured
- AWS CLI (recommended)

## Setup

1. Configure your AWS credentials:
   - Create `~/.aws/credentials` file or
   - Set environment variables: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

## Project Structure

```
.
├── cmd/                 # Command line applications
├── internal/            # Private application code
├── pkg/                 # Public library code
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── README.md           # This file
```
