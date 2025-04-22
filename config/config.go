package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Node              NodeConfig              `yaml:"node"`
	Secure            SecureConfig            `yaml:"secure"`
	SOCKS5            SOCKS5Config            `yaml:"socks5"`
	ForwardingPolicies []ForwardingPolicyConfig `yaml:"forwarding_policies"`
	Nodes             []RemoteNodeConfig      `yaml:"nodes"`
	HealthCheck       HealthCheckConfig       `yaml:"health_check"`
	Logging           LoggingConfig           `yaml:"logging"`
	Metrics           MetricsConfig           `yaml:"metrics"`
}

type NodeConfig struct {
	ID            string `yaml:"id"`
	ListenAddress string `yaml:"listen_address"`
	AdminAddress  string `yaml:"admin_address"`
}

type SecureConfig struct {
	Method    string    `yaml:"method"` // "tls", "aes", or "none"
	TLS       TLSConfig `yaml:"tls"`
	AES       AESConfig `yaml:"aes"`
}

type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

type AESConfig struct {
	KeyFile string `yaml:"key_file"` // File containing the AES key
}

type SOCKS5Config struct {
	Enabled bool        `yaml:"enabled"`
	Auth    SOCKS5Auth  `yaml:"auth"`
}

type SOCKS5Auth struct {
	Method          string `yaml:"method"`
	CredentialsFile string `yaml:"credentials_file"`
}

type ForwardingPolicyConfig struct {
	Name   string                 `yaml:"name"`
	Match  MatchConfig            `yaml:"match"`
	Action ForwardingActionConfig `yaml:"action"`
}

type MatchConfig struct {
	SourceCIDR      string `yaml:"source_cidr"`
	DestinationCIDR string `yaml:"destination_cidr"`
}

type ForwardingActionConfig struct {
	Type    string `yaml:"type"` // "direct" or "forward"
	NextHop string `yaml:"next_hop"`
}

type RemoteNodeConfig struct {
	ID        string `yaml:"id"`
	Address   string `yaml:"address"`
	PublicKey string `yaml:"public_key"`
}

type HealthCheckConfig struct {
	Interval          time.Duration `yaml:"interval"`
	Timeout           time.Duration `yaml:"timeout"`
	FailureThreshold  int           `yaml:"failure_threshold"`
	RecoveryThreshold int           `yaml:"recovery_threshold"`
	Method            string        `yaml:"method"` // "tcp_connect" or "application"
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefaults(&config)

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults(config *Config) {
	if config.Node.ListenAddress == "" {
		config.Node.ListenAddress = "0.0.0.0:1080"
	}

	if config.Secure.Method == "" {
		config.Secure.Method = "tls" // Default to TLS for backward compatibility
	}

	if config.HealthCheck.Interval == 0 {
		config.HealthCheck.Interval = 5 * time.Second
	}
	if config.HealthCheck.Timeout == 0 {
		config.HealthCheck.Timeout = 2 * time.Second
	}
	if config.HealthCheck.FailureThreshold == 0 {
		config.HealthCheck.FailureThreshold = 3
	}
	if config.HealthCheck.RecoveryThreshold == 0 {
		config.HealthCheck.RecoveryThreshold = 2
	}
	if config.HealthCheck.Method == "" {
		config.HealthCheck.Method = "tcp_connect"
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	if config.Metrics.Enabled && config.Metrics.Address == "" {
		config.Metrics.Address = "127.0.0.1:9090"
	}
}

func validate(config *Config) error {
	if config.Node.ID == "" {
		return fmt.Errorf("node.id is required")
	}

	if len(config.Nodes) > 0 {
		switch config.Secure.Method {
		case "tls":
			if config.Secure.TLS.CertFile == "" {
				return fmt.Errorf("secure.tls.cert_file is required when method is 'tls'")
			}
			if config.Secure.TLS.KeyFile == "" {
				return fmt.Errorf("secure.tls.key_file is required when method is 'tls'")
			}
			if config.Secure.TLS.CAFile == "" {
				return fmt.Errorf("secure.tls.ca_file is required when method is 'tls'")
			}
		case "aes":
			if config.Secure.AES.KeyFile == "" {
				return fmt.Errorf("secure.aes.key_file is required when method is 'aes'")
			}
		case "none":
		default:
			return fmt.Errorf("secure.method must be 'tls', 'aes', or 'none'")
		}
	}

	if config.SOCKS5.Enabled {
		if config.SOCKS5.Auth.Method != "none" && config.SOCKS5.Auth.Method != "username_password" {
			return fmt.Errorf("socks5.auth.method must be 'none' or 'username_password'")
		}
		if config.SOCKS5.Auth.Method == "username_password" && config.SOCKS5.Auth.CredentialsFile == "" {
			return fmt.Errorf("socks5.auth.credentials_file is required when auth.method is 'username_password'")
		}
	}

	if config.HealthCheck.Method != "tcp_connect" && config.HealthCheck.Method != "application" {
		return fmt.Errorf("health_check.method must be 'tcp_connect' or 'application'")
	}

	return nil
}
