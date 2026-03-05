# vpsctl - VPS Management CLI with API Backend

## Project Overview

Modern VPS management tool using LXD, written in Go. Provides CLI, API server, and TUI dashboard in a single binary.

## Architecture

```
vpsctl/
├── cmd/
│   ├── root.go              # Root command (cobra)
│   ├── create.go            # vpsctl create
│   ├── list.go              # vpsctl list
│   ├── start.go             # vpsctl start
│   ├── stop.go              # vpsctl stop
│   ├── restart.go           # vpsctl restart
│   ├── delete.go            # vpsctl delete
│   ├── shell.go             # vpsctl shell
│   ├── resize.go            # vpsctl resize
│   ├── snapshot.go          # vpsctl snapshot
│   ├── serve.go             # vpsctl serve (API mode)
│   └── dashboard.go         # vpsctl dashboard (TUI)
├── internal/
│   ├── lxd/
│   │   ├── client.go        # LXD client wrapper
│   │   ├── instance.go      # Instance operations
│   │   ├── network.go       # Network operations
│   │   ├── storage.go       # Storage operations
│   │   └── snapshot.go      # Snapshot operations
│   ├── resource/
│   │   ├── tracker.go       # Resource tracking
│   │   └── validator.go     # Resource validation
│   ├── portforward/
│   │   └── manager.go       # Port forwarding logic
│   └── config/
│       └── config.go        # Configuration management
├── api/
│   ├── server.go            # HTTP server setup
│   ├── handlers/
│   │   ├── instance.go      # Instance API handlers
│   │   ├── resource.go      # Resource API handlers
│   │   ├── image.go         # Image API handlers
│   │   └── websocket.go     # WebSocket handlers
│   ├── middleware/
│   │   ├── auth.go          # Authentication middleware
│   │   └── logging.go       # Request logging
│   └── routes.go            # Route definitions
├── tui/
│   ├── app.go               # Bubbletea app
│   ├── components/
│   │   ├── dashboard.go     # Main dashboard view
│   │   ├── instance_list.go # Instance list component
│   │   ├── resource_bar.go  # Resource usage bars
│   │   └── create_form.go   # Create VPS form
│   └── styles.go            # Lipgloss styles
├── pkg/
│   ├── output/
│   │   ├── table.go         # Table output formatter
│   │   └── json.go          # JSON output formatter
│   └── utils/
│       ├── size.go          # Size parsing (GB, MB, etc)
│       └── validate.go      # Input validation
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── main.go                  # Entry point
```

## Tech Stack

| Component | Package | Version |
|-----------|---------|---------|
| CLI Framework | github.com/spf13/cobra | v1.8.x |
| CLI Flags | github.com/spf13/pflag | v1.0.x |
| Config | github.com/spf13/viper | v1.18.x |
| TUI | github.com/charmbracelet/bubbletea | v0.25.x |
| TUI Components | github.com/charmbracelet/bubbles | v0.17.x |
| Styling | github.com/charmbracelet/lipgloss | v0.9.x |
| HTTP Router | github.com/gin-gonic/gin | v1.9.x |
| LXD Client | github.com/canonical/lxd | v0.0.0 (latest) |
| WebSocket | github.com/gorilla/websocket | v1.5.x |
| Logging | github.com/sirupsen/logrus | v1.9.x |
| Validation | github.com/go-playground/validator | v10.x |

## API Endpoints

### Instances
- `GET /api/v1/instances` - List all instances
- `POST /api/v1/instances` - Create new instance
- `GET /api/v1/instances/:name` - Get instance details
- `PUT /api/v1/instances/:name` - Update instance
- `DELETE /api/v1/instances/:name` - Delete instance
- `POST /api/v1/instances/:name/start` - Start instance
- `POST /api/v1/instances/:name/stop` - Stop instance
- `POST /api/v1/instances/:name/restart` - Restart instance
- `GET /api/v1/instances/:name/metrics` - Get instance metrics (WebSocket)

### Resources
- `GET /api/v1/resources` - Host resource summary
- `GET /api/v1/resources/allocated` - Allocated resources

### Images
- `GET /api/v1/images` - List available images

### Port Forwarding
- `GET /api/v1/instances/:name/ports` - List port forwards
- `POST /api/v1/instances/:name/ports` - Add port forward
- `DELETE /api/v1/instances/:name/ports/:port` - Remove port forward

## CLI Commands

```bash
# Instance Management
vpsctl create <name> [flags]
  --image, -i string     Image to use (default "ubuntu:24.04")
  --cpu, -c int          CPU cores (default 1)
  --memory, -m string    Memory limit (default "512MB")
  --disk, -d string      Disk size (default "10GB")
  --type, -t string      Instance type: container or vm (default "container")
  --ssh-key string       SSH public key
  --password string      Root password

vpsctl list [flags]
  --format, -f string    Output format: table, json, csv (default "table")
  --all, -a              Show all details

vpsctl start <name>
vpsctl stop <name> [--force]
vpsctl restart <name>
vpsctl delete <name> [--force]

vpsctl shell <name> [--user, -u string]  # Exec into instance
vpsctl resize <name> [flags]
  --cpu int              New CPU limit
  --memory string        New memory limit
  --disk string          New disk size

vpsctl snapshot <name> [flags]
  --name, -n string      Snapshot name
  --list                 List snapshots
  --restore string       Restore from snapshot

# API Server
vpsctl serve [flags]
  --port, -p int         API port (default 8080)
  --socket string        Unix socket path
  --auth                 Enable authentication
  --token string         API token

# TUI Dashboard
vpsctl dashboard

# Resource Info
vpsctl resources
```

## Parallel Development Tasks

### Agent 1: Core Infrastructure (Project Setup + LXD Client)
- Initialize Go module
- Create project structure
- Implement LXD client wrapper
- Implement instance operations (create, list, start, stop, delete)

### Agent 2: CLI Commands (Cobra Layer)
- Implement all CLI commands
- Flag parsing
- Output formatting (table, json)
- Input validation

### Agent 3: API Server
- HTTP server setup with Gin
- All API handlers
- Middleware (logging, auth)
- WebSocket for real-time metrics

### Agent 4: TUI Dashboard
- Bubbletea application
- Dashboard components
- Create VPS form
- Resource visualization

### Agent 5: Utilities & Config
- Size parsing utilities
- Resource tracker
- Configuration management
- Makefile + README

## File Assignments

| Agent | Files |
|-------|-------|
| Agent 1 | `go.mod`, `main.go`, `internal/lxd/*`, `pkg/utils/size.go` |
| Agent 2 | `cmd/*.go`, `pkg/output/*`, `pkg/utils/validate.go` |
| Agent 3 | `api/*`, `internal/portforward/*` |
| Agent 4 | `tui/*` |
| Agent 5 | `internal/config/*`, `internal/resource/*`, `Makefile`, `README.md` |

## Dependencies Graph

```
Agent 1 (Core) ─────┬──► Agent 2 (CLI)
                    ├──► Agent 3 (API)
                    ├──► Agent 4 (TUI)
                    └──► Agent 5 (Config)
```

Agent 1 must complete first, then Agents 2-5 can run in parallel.

## Build Commands

```bash
# Development
make dev

# Build
make build

# Build for all platforms
make build-all

# Run tests
make test

# Install
make install
```

## Progress Tracking

- [ ] Phase 1: Project Setup & Core Library
- [ ] Phase 2: CLI Commands
- [ ] Phase 3: API Server
- [ ] Phase 4: TUI Dashboard
- [ ] Phase 5: Utilities & Documentation
- [ ] Phase 6: Integration & Testing
