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

## Usage Scenarios

### Browser with SOCKS5 Plugin and AES Encryption

This scenario demonstrates how to configure Water Pipe for a browser using a SOCKS5 plugin to access the internet through the proxy, with node-to-node communication encrypted using AES.

#### Node Configuration

```yaml
node:
  id: "node1"
  listen_address: "0.0.0.0:1080"  # Listen on all interfaces on port 1080
  admin_address: "127.0.0.1:8080" # Admin interface on localhost

secure:
  method: "aes"  # Use AES encryption for node-to-node communication
  aes:
    key_file: "/etc/water_pipe/aes.key"  # Path to AES key file

socks5:
  enabled: true  # Enable SOCKS5 server
  auth:
    method: "none"  # No authentication required (or use "username_password")

forwarding_policies:
  - name: "direct_policy"
    match:
      destination_cidr: "192.168.0.0/16"  # Local network traffic
    action:
      type: "direct"  # Direct connection
  
  - name: "next_hop_policy"
    match:
      destination_cidr: "0.0.0.0/0"  # All other traffic
    action:
      type: "forward"  # Forward to next hop
      next_hop: "node2"  # ID of the next node

nodes:
  - id: "node2"
    address: "proxy2.example.com:443"  # Address of the next node
    public_key: "/etc/water_pipe/node2_pubkey.pem"  # Public key of the next node

health_check:
  interval: 5s  # Check every 5 seconds
  timeout: 2s   # Timeout after 2 seconds
  failure_threshold: 3  # Mark as unhealthy after 3 failures
  recovery_threshold: 2  # Mark as healthy after 2 successes
  method: "tcp_connect"  # Use TCP connect for health checks
```

#### Browser Configuration

1. Install a SOCKS5 proxy plugin for your browser (e.g., SwitchyOmega for Chrome/Firefox)
2. Configure the plugin to use the proxy at `127.0.0.1:1080` (or the IP address of your proxy node)
3. Set the proxy type to SOCKS5
4. Save and apply the configuration

Now your browser traffic will be routed through the Water Pipe proxy network, with node-to-node communication encrypted using AES.

### Node as a Next Hop Without SOCKS5

You can configure a node as a next hop for other nodes without enabling SOCKS5. The node will still listen on the configured `listen_address` and accept connections from other nodes in the network.

Example configuration for a node without SOCKS5 that serves as a next hop:

```yaml
node:
  id: "relay-node"
  listen_address: "0.0.0.0:1081"  # Listen on all interfaces on port 1081

secure:
  method: "aes"  # Use AES encryption for node-to-node communication
  aes:
    key_file: "/etc/water_pipe/aes.key"  # Path to AES key file

socks5:
  enabled: false  # SOCKS5 is disabled

# The rest of the configuration remains the same
```

## Documentation

- [Configuration Guide](docs/configuration.md)
- [Deployment Guide](docs/deployment.md)
- [Security Guide](docs/security.md)
- [API Reference](docs/api.md)

## License

[MIT License](LICENSE)
