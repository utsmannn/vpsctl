# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release

## [1.0.0] - 2024-01-XX

### Added
- CLI commands for VPS management
  - `create` - Create new VPS instances
  - `list` - List all instances
  - `start` - Start instances
  - `stop` - Stop instances
  - `restart` - Restart instances
  - `delete` - Delete instances
  - `shell` - Access instance shell
  - `resize` - Resize instance resources
  - `snapshot` - Manage instance snapshots
  - `resources` - View resource summary
- REST API server with endpoints for:
  - Instance CRUD operations
  - Resource monitoring
  - WebSocket real-time metrics
  - Image listing
  - Port forwarding
- TUI Dashboard with:
  - Instance list with status
  - Resource visualization
  - Quick actions
- Multi-platform support (Linux, macOS)
- Systemd service integration
- Install script for easy deployment

### Security
- API token authentication support
- CORS configuration

[Unreleased]: https://github.com/utsmannn/vpsctl/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/utsmannn/vpsctl/releases/tag/v1.0.0
