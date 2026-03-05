# vpsctl

Modern VPS Management CLI with API Backend powered by LXD.

## Features

- **CLI Interface** - Full-featured command line interface for VPS management
- **REST API** - Built-in API server for dashboard integration
- **TUI Dashboard** - Interactive terminal dashboard for monitoring
- **Real-time Metrics** - WebSocket-based live monitoring
- **Resource Tracking** - CPU, RAM, and disk allocation tracking
- **Port Forwarding** - Automatic port detection and forwarding

## Installation

### One-line Install (Recommended)

```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/utsmannn/vpsctl/main/scripts/install.sh | bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/utsmannn/vpsctl/main/scripts/install.sh | bash -s -- -v v1.0.0

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/utsmannn/vpsctl/main/scripts/install.sh | bash -s -- -b ~/bin
```

### Manual Download

Download the binary for your platform from [Releases](https://github.com/utsmannn/vpsctl/releases):

```bash
# Linux AMD64
curl -Lo vpsctl https://github.com/utsmannn/vpsctl/releases/latest/download/vpsctl-linux-amd64
chmod +x vpsctl
sudo mv vpsctl /usr/local/bin/

# Linux ARM64
curl -Lo vpsctl https://github.com/utsmannn/vpsctl/releases/latest/download/vpsctl-linux-arm64
chmod +x vpsctl
sudo mv vpsctl /usr/local/bin/

# macOS (Apple Silicon)
curl -Lo vpsctl https://github.com/utsmannn/vpsctl/releases/latest/download/vpsctl-darwin-arm64
chmod +x vpsctl
sudo mv vpsctl /usr/local/bin/

# macOS (Intel)
curl -Lo vpsctl https://github.com/utsmannn/vpsctl/releases/latest/download/vpsctl-darwin-amd64
chmod +x vpsctl
sudo mv vpsctl /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/utsmannn/vpsctl.git
cd vpsctl
make install
```

### Requirements

- Go 1.21+ (for building from source)
- LXD installed and initialized
- User in `lxd` group

---

## Deployment

### GitHub Actions CI/CD

The project includes a GitHub Actions workflow that automatically:
- Builds binaries for multiple platforms (Linux, macOS)
- Creates GitHub releases with checksums
- Optionally deploys to your server

**Creating a Release:**

```bash
# Tag and push
git tag v1.0.0
git push origin v1.0.0

# Or trigger manually via GitHub UI (Actions > Build and Release > Run workflow)
```

### Installing on Target Server

**Option 1: Using Install Script**

```bash
# On target server
curl -fsSL https://raw.githubusercontent.com/utsmannn/vpsctl/main/scripts/install.sh | bash
```

**Option 2: Manual Installation**

```bash
# Download
curl -Lo vpsctl https://github.com/utsmannn/vpsctl/releases/latest/download/vpsctl-linux-amd64
chmod +x vpsctl
sudo mv vpsctl /usr/local/bin/

# Verify
vpsctl --version
```

### Running as Systemd Service

```bash
# Copy service file
sudo curl -Lo /etc/systemd/system/vpsctl.service \
  https://raw.githubusercontent.com/utsmannn/vpsctl/main/scripts/vpsctl.service

# Edit if needed (port, token, etc.)
sudo nano /etc/systemd/system/vpsctl.service

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable vpsctl
sudo systemctl start vpsctl

# Check status
sudo systemctl status vpsctl
```

### Auto-Deploy via GitHub Actions

To enable automatic deployment to your server on release:

1. Add these secrets to your GitHub repository:
   - `DEPLOY_HOST` - Your server IP or hostname
   - `DEPLOY_USER` - SSH username (e.g., `root` or `ubuntu`)
   - `DEPLOY_KEY` - SSH private key (contents of your `.pem` file)

2. Create a GitHub environment named `production`

3. Push a new tag to trigger deployment:

```bash
git tag v1.0.1
git push origin v1.0.1
```

### Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/vpsctl

# Remove systemd service (if installed)
sudo systemctl stop vpsctl
sudo systemctl disable vpsctl
sudo rm /etc/systemd/system/vpsctl.service
sudo systemctl daemon-reload
```

## Quick Start

```bash
# Create a new VPS
vpsctl create my-server --image ubuntu:24.04 --cpu 2 --memory 1GB --disk 10GB

# List all VPS
vpsctl list

# Start a VPS
vpsctl start my-server

# Access shell
vpsctl shell my-server

# Start API server
vpsctl serve --port 8080

# Launch TUI dashboard
vpsctl dashboard

# Show resource summary
vpsctl resources
```

## CLI Commands

### Instance Management

| Command | Description |
|---------|-------------|
| `vpsctl create <name>` | Create new VPS instance |
| `vpsctl list` | List all instances |
| `vpsctl start <name>` | Start instance |
| `vpsctl stop <name>` | Stop instance |
| `vpsctl restart <name>` | Restart instance |
| `vpsctl delete <name>` | Delete instance |
| `vpsctl shell <name>` | Exec into instance |
| `vpsctl resize <name>` | Resize instance resources |
| `vpsctl snapshot <name>` | Manage snapshots |

### Create Command Options

```bash
vpsctl create <name> [flags]

Flags:
  --image, -i string     Image to use (default "ubuntu:24.04")
  --cpu, -c int          CPU cores (default 1)
  --memory, -m string    Memory limit (default "512MB")
  --disk, -d string      Disk size (default "10GB")
  --type, -t string      Instance type: container or vm (default "container")
  --ssh-key string       SSH public key for access
  --password string      Root password
```

### Server & Dashboard

| Command | Description |
|---------|-------------|
| `vpsctl serve` | Start API server |
| `vpsctl dashboard` | Launch TUI dashboard |
| `vpsctl resources` | Show resource summary |

### API Server Options

```bash
vpsctl serve [flags]

Flags:
  --port, -p int         API port (default 8080)
  --socket string        Unix socket path (optional)
  --auth                 Enable authentication
  --token string         API token for authentication
```

## API Endpoints

### Instances

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/instances` | List all instances |
| POST | `/api/v1/instances` | Create instance |
| GET | `/api/v1/instances/:name` | Get instance details |
| PUT | `/api/v1/instances/:name` | Update instance |
| DELETE | `/api/v1/instances/:name` | Delete instance |
| POST | `/api/v1/instances/:name/start` | Start instance |
| POST | `/api/v1/instances/:name/stop` | Stop instance |
| POST | `/api/v1/instances/:name/restart` | Restart instance |
| GET | `/api/v1/instances/:name/metrics` | Get instance metrics (WebSocket) |

### Resources

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/resources` | Host resource summary |
| GET | `/api/v1/resources/allocated` | Allocated resources |

### Images

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/images` | List available images |

### Port Forwarding

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/instances/:name/ports` | List port forwards |
| POST | `/api/v1/instances/:name/ports` | Add port forward |
| DELETE | `/api/v1/instances/:name/ports/:port` | Remove port forward |

## Configuration

Config file location: `~/.vpsctl.yaml`

```yaml
lxd:
  socket: /var/snap/lxd/common/lxd/unix.socket

api:
  port: 8080
  auth: false
  token: ""
  cors: true

output:
  format: table
```

### Environment Variables

Configuration can also be set via environment variables with the `VPSCTL_` prefix:

```bash
export VPSCTL_LXD_SOCKET=/var/snap/lxd/common/lxd/unix.socket
export VPSCTL_API_PORT=8080
export VPSCTL_OUTPUT_FORMAT=json
```

## Development

### Build Commands

```bash
# Install dependencies
make deps

# Development run
make dev

# Build binary
make build

# Run tests
make test

# Run all checks (fmt, vet, test)
make check

# Build for all platforms
make build-all

# Create release
make release VERSION=0.2.0
```

### Cross-Compilation

```bash
# Build for Linux
make build-linux

# Build for macOS (Intel and Apple Silicon)
make build-darwin

# Build for Windows
make build-windows

# Build for all platforms
make build-all
```

### Docker

```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

## Project Structure

```
vpsctl/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go
│   ├── create.go
│   ├── list.go
│   ├── start.go
│   ├── stop.go
│   ├── restart.go
│   ├── delete.go
│   ├── shell.go
│   ├── resize.go
│   ├── snapshot.go
│   ├── serve.go
│   └── dashboard.go
├── internal/               # Private application code
│   ├── lxd/               # LXD client wrapper
│   ├── config/            # Configuration management
│   ├── resource/          # Resource tracking & validation
│   └── portforward/       # Port forwarding logic
├── api/                    # API server
│   ├── server.go
│   ├── handlers/
│   ├── middleware/
│   └── routes.go
├── tui/                    # Terminal UI (Bubbletea)
│   ├── app.go
│   ├── components/
│   └── styles.go
├── pkg/                    # Public packages
│   ├── output/            # Output formatters
│   └── utils/             # Utility functions
├── main.go                 # Entry point
├── Makefile
├── go.mod
└── README.md
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
| LXD Client | github.com/canonical/lxd | latest |
| WebSocket | github.com/gorilla/websocket | v1.5.x |
| Logging | github.com/sirupsen/logrus | v1.9.x |
| Validation | github.com/go-playground/validator | v10.x |

## Examples

### Creating a VPS

```bash
# Basic container
vpsctl create web-server

# With custom resources
vpsctl create db-server --cpu 4 --memory 4GB --disk 50GB

# Using a specific image
vpsctl create app-server --image ubuntu:22.04 --cpu 2 --memory 2GB

# As a virtual machine
vpsctl create vm-server --type vm --cpu 2 --memory 4GB --disk 20GB
```

### Managing Instances

```bash
# List all instances
vpsctl list

# List in JSON format
vpsctl list --format json

# Start/Stop/Restart
vpsctl start web-server
vpsctl stop web-server
vpsctl restart web-server

# Force stop
vpsctl stop web-server --force

# Delete an instance
vpsctl delete web-server

# Force delete (stops first)
vpsctl delete web-server --force
```

### Shell Access

```bash
# Access as root
vpsctl shell web-server

# Access as specific user
vpsctl shell web-server --user ubuntu
```

### Resource Management

```bash
# View resource summary
vpsctl resources

# Resize instance
vpsctl resize web-server --cpu 4
vpsctl resize web-server --memory 2GB
vpsctl resize web-server --disk 20GB
```

### Snapshots

```bash
# Create snapshot
vpsctl snapshot web-server --name backup-2024-01-15

# List snapshots
vpsctl snapshot web-server --list

# Restore from snapshot
vpsctl snapshot web-server --restore backup-2024-01-15

# Delete snapshot
vpsctl snapshot web-server --delete backup-2024-01-15
```

### API Server

```bash
# Start API server
vpsctl serve

# Start with authentication
vpsctl serve --auth --token my-secret-token

# Start on custom port
vpsctl serve --port 3000

# Start with Unix socket
vpsctl serve --socket /tmp/vpsctl.sock
```

## Troubleshooting

### LXD Permission Issues

If you get permission errors connecting to LXD:

```bash
# Add your user to the lxd group
sudo usermod -aG lxd $USER

# Log out and back in, or run:
newgrp lxd

# Verify LXD is working
lxc list
```

### Socket Path Issues

If using a non-standard LXD socket:

```bash
# Set via config
vpsctl config set lxd.socket /path/to/socket

# Or via environment variable
export VPSCTL_LXD_SOCKET=/path/to/socket
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- GitHub Issues: [https://github.com/utsmannn/vpsctl/issues](https://github.com/utsmannn/vpsctl/issues)
- Documentation: [https://github.com/utsmannn/vpsctl/wiki](https://github.com/utsmannn/vpsctl/wiki)
