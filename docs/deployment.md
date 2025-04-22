# Deployment Guide

This guide provides instructions for deploying Water Pipe on both Linux and Windows systems.

## Prerequisites

- Go 1.20 or higher
- TLS certificates for secure node-to-node communication

## Building from Source

### Linux

1. Clone the repository:
   ```bash
   git clone https://github.com/a465749927/water_pipe.git
   cd water_pipe
   ```

2. Build the binary:
   ```bash
   make build
   ```

3. Alternatively, build for a specific platform:
   ```bash
   make build-linux    # For Linux
   make build-windows  # For Windows
   ```

4. The binary will be created in the current directory.

### Windows

1. Clone the repository:
   ```powershell
   git clone https://github.com/a465749927/water_pipe.git
   cd water_pipe
   ```

2. Build the binary using the PowerShell script:
   ```powershell
   .\build.ps1
   ```

3. The binary will be created in the current directory.

## Configuration

1. Create a configuration file based on the example in the [Configuration Guide](configuration.md).

2. Save the configuration file as `config.yaml` in the same directory as the binary, or specify a different path using the `-config` flag.

## TLS Certificate Setup

For secure node-to-node communication, you need to set up TLS certificates:

1. Generate a CA certificate:
   ```bash
   openssl genrsa -out ca.key 4096
   openssl req -new -x509 -key ca.key -sha256 -subj "/CN=Water Pipe CA" -out ca.pem -days 365
   ```

2. Generate a certificate for each node:
   ```bash
   # Generate private key
   openssl genrsa -out node1.key 2048
   
   # Create certificate signing request
   openssl req -new -key node1.key -out node1.csr -subj "/CN=node1"
   
   # Sign the certificate
   openssl x509 -req -in node1.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out node1.pem -days 365 -sha256
   ```

3. Update the configuration file with the paths to the certificates.

## Running the Proxy

### Linux

```bash
# Run with a specific configuration file
./water_pipe -config /path/to/config.yaml

# Run with default configuration (config.yaml in the current directory)
./water_pipe
```

### Windows

```powershell
# Run with a specific configuration file
.\water_pipe.exe -config C:\path\to\config.yaml

# Run with default configuration (config.yaml in the current directory)
.\water_pipe.exe
```

## Running as a Service

### Linux (systemd)

1. Create a systemd service file:
   ```bash
   sudo nano /etc/systemd/system/water_pipe.service
   ```

2. Add the following content:
   ```
   [Unit]
   Description=Water Pipe Proxy
   After=network.target

   [Service]
   ExecStart=/path/to/water_pipe -config /path/to/config.yaml
   Restart=on-failure
   User=nobody
   Group=nogroup

   [Install]
   WantedBy=multi-user.target
   ```

3. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable water_pipe
   sudo systemctl start water_pipe
   ```

4. Check the status:
   ```bash
   sudo systemctl status water_pipe
   ```

### Windows (Windows Service)

1. Install [NSSM (Non-Sucking Service Manager)](https://nssm.cc/):
   ```powershell
   # Using Chocolatey
   choco install nssm
   ```

2. Create a service:
   ```powershell
   nssm install WaterPipe
   ```

3. In the NSSM dialog:
   - Set the path to the water_pipe.exe
   - Set the startup directory
   - Add arguments: `-config C:\path\to\config.yaml`
   - Configure other options as needed

4. Start the service:
   ```powershell
   nssm start WaterPipe
   ```

## Network Configuration

### Firewall Configuration

Make sure to open the necessary ports in your firewall:

- SOCKS5 server port (default: 1080)
- Node-to-node communication port (typically 443)
- Admin API port (if enabled)

#### Linux (iptables)

```bash
# Allow SOCKS5 server port
sudo iptables -A INPUT -p tcp --dport 1080 -j ACCEPT

# Allow node-to-node communication port
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

#### Windows (Windows Firewall)

```powershell
# Allow SOCKS5 server port
New-NetFirewallRule -DisplayName "Water Pipe SOCKS5" -Direction Inbound -Protocol TCP -LocalPort 1080 -Action Allow

# Allow node-to-node communication port
New-NetFirewallRule -DisplayName "Water Pipe Node Communication" -Direction Inbound -Protocol TCP -LocalPort 443 -Action Allow
```

## Monitoring

Water Pipe provides metrics in Prometheus format if enabled in the configuration. You can use Prometheus and Grafana to monitor the proxy:

1. Configure Prometheus to scrape metrics from the Water Pipe metrics endpoint.

2. Import the Water Pipe dashboard into Grafana.

## Troubleshooting

### Common Issues

1. **Connection refused**: Check if the proxy is running and listening on the correct address.

2. **TLS handshake failed**: Verify that the certificates are correctly configured and accessible.

3. **Authentication failed**: Check the credentials file and authentication method configuration.

### Logs

Check the logs for error messages:

- If output is set to stdout/stderr, check the console output or service logs.
- If output is set to a file, check the specified log file.

### Debugging

Run the proxy with increased log level for more detailed information:

```yaml
logging:
  level: "debug"
```

## Security Considerations

1. **TLS Certificates**: Keep private keys secure and rotate certificates regularly.

2. **Authentication**: Use strong passwords for SOCKS5 authentication.

3. **Firewall Rules**: Restrict access to the proxy to trusted networks.

4. **Regular Updates**: Keep the proxy software updated with the latest security patches.
