# Security Guide

This guide provides information about the security features of Water Pipe and best practices for secure deployment.

## Security Features

### Encrypted Node-to-Node Communication

Water Pipe uses TLS 1.3 for all node-to-node communication, providing:

- **Confidentiality**: All traffic between nodes is encrypted.
- **Integrity**: Any tampering with the traffic is detected.
- **Authentication**: Nodes authenticate each other using X.509 certificates.

### SOCKS5 Authentication

Water Pipe supports SOCKS5 authentication methods:

- **No Authentication**: Suitable for trusted networks.
- **Username/Password Authentication**: Provides basic access control.

### Policy-Based Routing

Traffic can be routed based on policies, allowing for:

- **Network Segmentation**: Different traffic can be routed through different paths.
- **Access Control**: Traffic from certain sources can be restricted.

## Security Best Practices

### Certificate Management

1. **Use Strong Keys**: Use RSA keys with at least 2048 bits or ECC keys with at least 256 bits.

2. **Protect Private Keys**: Store private keys securely with appropriate file permissions.

3. **Certificate Rotation**: Regularly rotate certificates to limit the impact of key compromise.

4. **Certificate Validation**: Always validate certificates against a trusted CA.

### Network Security

1. **Firewall Configuration**: Restrict access to proxy nodes to trusted networks.

2. **Network Segmentation**: Deploy proxy nodes in different network segments for better isolation.

3. **Monitoring**: Monitor traffic patterns for anomalies that might indicate compromise.

### Authentication

1. **Strong Passwords**: Use strong, unique passwords for SOCKS5 authentication.

2. **Credential Management**: Store credentials securely and rotate them regularly.

3. **Access Control**: Limit access to the proxy to authorized users only.

### Deployment

1. **Least Privilege**: Run the proxy with the minimum necessary privileges.

2. **Regular Updates**: Keep the proxy software and dependencies updated.

3. **Secure Configuration**: Review and harden the configuration file.

4. **Logging**: Enable appropriate logging to detect and investigate security incidents.

## TLS Configuration

### Recommended TLS Settings

Water Pipe uses TLS 1.3 by default, which provides strong security. However, you should ensure:

1. **Strong Certificates**: Use certificates from a trusted CA or properly managed internal CA.

2. **Certificate Verification**: Always enable certificate verification.

3. **Private Key Protection**: Protect private keys with appropriate file permissions:
   ```bash
   chmod 600 /path/to/key.pem
   ```

### Certificate Generation

For testing or internal use, you can generate self-signed certificates:

```bash
# Generate CA key and certificate
openssl genrsa -out ca.key 4096
openssl req -new -x509 -key ca.key -sha256 -subj "/CN=Water Pipe CA" -out ca.pem -days 365

# Generate node key and certificate signing request
openssl genrsa -out node.key 2048
openssl req -new -key node.key -out node.csr -subj "/CN=node.example.com"

# Sign the certificate
openssl x509 -req -in node.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out node.pem -days 365 -sha256
```

For production use, consider using a proper certificate management solution or a public CA.

## SOCKS5 Authentication

### Username/Password Authentication

When using username/password authentication, create a credentials file:

```json
[
  {
    "username": "user1",
    "password": "strong-password-1"
  },
  {
    "username": "user2",
    "password": "strong-password-2"
  }
]
```

Protect this file with appropriate permissions:

```bash
chmod 600 /path/to/credentials.json
```

## Secure Deployment Checklist

Before deploying Water Pipe in production, ensure:

1. **TLS is properly configured** with valid certificates.

2. **Authentication is enabled** if the proxy is accessible from untrusted networks.

3. **Firewall rules are in place** to restrict access to the proxy.

4. **Logging is configured** to capture security-relevant events.

5. **The proxy runs with minimal privileges** on the host system.

6. **All files (certificates, keys, credentials) have appropriate permissions**.

7. **The configuration file does not contain sensitive information** in plain text.

8. **Health checks are enabled** to detect and route around compromised nodes.

## Incident Response

If you suspect a security incident:

1. **Isolate the affected node** by removing it from the routing configuration.

2. **Revoke and replace certificates** if you suspect key compromise.

3. **Analyze logs** to understand the scope and nature of the incident.

4. **Update credentials** and configuration as necessary.

5. **Deploy a patched version** if the incident was due to a software vulnerability.

## Security Updates

Stay informed about security updates:

1. **Monitor the project repository** for security announcements.

2. **Subscribe to security mailing lists** for Go and any dependencies.

3. **Regularly check for updates** to the proxy software and dependencies.
