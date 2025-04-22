# Water Pipe

A high-performance, generic four-layer proxy written in Go, designed for deployment on both Linux and Windows.

## Features

- **Traffic Forwarding Policies**: Configure how traffic is routed through the proxy network
- **Secure Node-to-Node Communication**: All inter-node communication is encrypted using TLS 1.3
- **Node Health Probing**: Automatically detect and route around failed nodes
- **SOCKS5 Proxy Support**: Compatible with standard SOCKS5 clients
- **Cross-Platform**: Runs on both Linux and Windows

## Architecture

Water Pipe functions as a network of interconnected proxy nodes, each capable of making intelligent forwarding decisions based on configured policies and real-time health checks.

```
┌─────────────────────────────────────────────────────────────────┐
│                          Proxy Node                             │
│                                                                 │
│  ┌───────────────┐    ┌───────────────┐    ┌───────────────┐   │
│  │  SOCKS5 Server │    │ Traffic Router │    │ Health Checker│   │
│  └───────┬───────┘    └───────┬───────┘    └───────┬───────┘   │
│          │                    │                    │           │
│          └────────────┬───────┴────────────┬──────┘           │
│                       │                    │                   │
│               ┌───────┴───────┐    ┌──────┴───────┐           │
│               │ Secure Comms  │    │Configuration │           │
│               │    Layer      │    │   Manager    │           │
│               └───────────────┘    └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.20 or higher

### Building from Source

#### Linux

```bash
git clone https://github.com/a465749927/water_pipe.git
cd water_pipe
make build
```

#### Windows

```powershell
git clone https://github.com/a465749927/water_pipe.git
cd water_pipe
.\build.ps1
```

## Configuration

Water Pipe uses YAML for configuration. See the [configuration guide](docs/configuration.md) for details.

Example configuration:

```yaml
node:
  id: "node1"
  listen_address: "0.0.0.0:1080"

tls:
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  ca_file: "/path/to/ca.pem"

socks5:
  enabled: true
  auth:
    method: "none"

forwarding_policies:
  - name: "direct_policy"
    match:
      destination_cidr: "0.0.0.0/0"
    action:
      type: "direct"
```

## Usage

```bash
# Run with a specific configuration file
water_pipe -config /path/to/config.yaml

# Run with default configuration
water_pipe
```

## Documentation

- [Configuration Guide](docs/configuration.md)
- [Deployment Guide](docs/deployment.md)
- [Security Guide](docs/security.md)
- [API Reference](docs/api.md)

## License

[MIT License](LICENSE)
