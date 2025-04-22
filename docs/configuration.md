# Configuration Guide

Water Pipe uses YAML for configuration. This document describes the available configuration options.

## Configuration File Structure

The configuration file is structured into several sections:

```yaml
node:
  # Node configuration
tls:
  # TLS configuration
socks5:
  # SOCKS5 configuration
forwarding_policies:
  # Forwarding policies
nodes:
  # Remote nodes
health_check:
  # Health check configuration
logging:
  # Logging configuration
metrics:
  # Metrics configuration
```

## Node Configuration

```yaml
node:
  id: "node1"                # Unique identifier for this node
  listen_address: "0.0.0.0:1080"  # Address to listen on for incoming connections
  admin_address: "127.0.0.1:8080" # Address for admin API (optional)
```

## Secure Communication Configuration

```yaml
secure:
  method: "tls"              # Encryption method: "tls", "aes", or "none"
  
  # TLS configuration (used when method is "tls")
  tls:
    cert_file: "/path/to/cert.pem"  # Path to certificate file
    key_file: "/path/to/key.pem"    # Path to private key file
    ca_file: "/path/to/ca.pem"      # Path to CA certificate file
  
  # AES configuration (used when method is "aes")
  aes:
    key_file: "/path/to/aes.key"    # Path to AES key file
  
  # No additional configuration needed for method: "none"
```

### Encryption Methods

Water Pipe supports three encryption methods for node-to-node communication:

1. **TLS (Transport Layer Security)**: Provides strong encryption and mutual authentication using X.509 certificates. This is the most secure option but may have slightly higher overhead.

2. **AES (Advanced Encryption Standard)**: Provides efficient symmetric encryption using a pre-shared key. This option offers a good balance between security and performance.

3. **None (No Encryption)**: Transmits data without encryption. This option provides the highest performance but should only be used in trusted networks or for testing purposes.

### AES Key Generation

To generate an AES key for use with the "aes" encryption method:

```bash
# Generate a 256-bit (32-byte) AES key
openssl rand -out aes.key 32
```

The AES key must be 16, 24, or 32 bytes in length, corresponding to AES-128, AES-192, or AES-256 respectively.

## SOCKS5 Configuration

```yaml
socks5:
  enabled: true              # Whether to enable SOCKS5 server
  auth:
    method: "none"           # Authentication method: "none" or "username_password"
    credentials_file: "/path/to/credentials.json"  # Path to credentials file (required if method is "username_password")
```

The credentials file should be a JSON file with the following structure:

```json
[
  {
    "username": "user1",
    "password": "password1"
  },
  {
    "username": "user2",
    "password": "password2"
  }
]
```

## Forwarding Policies

```yaml
forwarding_policies:
  - name: "direct_policy"    # Policy name
    match:
      source_cidr: "192.168.1.0/24"       # Source CIDR to match (optional)
      destination_cidr: "10.0.0.0/8"      # Destination CIDR to match (optional)
    action:
      type: "direct"         # Action type: "direct" or "forward"
  
  - name: "next_hop_policy"
    match:
      destination_cidr: "0.0.0.0/0"
    action:
      type: "forward"        # Forward to another node
      next_hop: "node2"      # Node ID to forward to
```

## Remote Nodes

```yaml
nodes:
  - id: "node2"              # Node ID
    address: "node2.example.com:443"  # Node address
    public_key: "/path/to/node2_pubkey.pem"  # Path to node's public key
  
  - id: "node3"
    address: "node3.example.com:443"
    public_key: "/path/to/node3_pubkey.pem"
```

## Health Check Configuration

```yaml
health_check:
  interval: 5s               # Interval between health checks
  timeout: 2s                # Timeout for health checks
  failure_threshold: 3       # Number of consecutive failures before marking a node as unhealthy
  recovery_threshold: 2      # Number of consecutive successes before marking a node as healthy
  method: "tcp_connect"      # Health check method: "tcp_connect" or "application"
```

## Logging Configuration

```yaml
logging:
  level: "info"              # Log level: "debug", "info", "warning", "error", "fatal"
  format: "text"             # Log format: "text" or "json"
  output: "stdout"           # Log output: "stdout", "stderr", or a file path
```

## Metrics Configuration

```yaml
metrics:
  enabled: true              # Whether to enable metrics
  address: "127.0.0.1:9090"  # Address to expose metrics on
```

## Example Configuration

```yaml
node:
  id: "node1"
  listen_address: "0.0.0.0:1080"

tls:
  cert_file: "/etc/water_pipe/cert.pem"
  key_file: "/etc/water_pipe/key.pem"
  ca_file: "/etc/water_pipe/ca.pem"

socks5:
  enabled: true
  auth:
    method: "none"

forwarding_policies:
  - name: "internal_network"
    match:
      destination_cidr: "10.0.0.0/8"
    action:
      type: "direct"
  
  - name: "default_policy"
    match:
      destination_cidr: "0.0.0.0/0"
    action:
      type: "forward"
      next_hop: "node2"

nodes:
  - id: "node2"
    address: "node2.example.com:443"
    public_key: "/etc/water_pipe/node2_pubkey.pem"

health_check:
  interval: 5s
  timeout: 2s
  failure_threshold: 3
  recovery_threshold: 2
  method: "tcp_connect"

logging:
  level: "info"
  format: "text"
  output: "stdout"

metrics:
  enabled: true
  address: "127.0.0.1:9090"
```
