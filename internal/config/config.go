package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete Reflect configuration.
type Config struct {
	// Environments defines named upstream environments for "Try It" functionality.
	Environments []Environment `yaml:"environments"`

	// HeaderAllowlist specifies which HTTP headers can be sent to upstream services.
	// This prevents accidentally leaking sensitive headers.
	HeaderAllowlist []string `yaml:"headerAllowlist"`

	// MaxRequestBodyBytes limits the size of request bodies for "Try It" invocations.
	// Default: 1048576 (1 MB).
	MaxRequestBodyBytes int64 `yaml:"maxRequestBodyBytes"`

	// RequestTimeoutSeconds sets the timeout for upstream RPC calls.
	// Default: 15 seconds.
	RequestTimeoutSeconds int `yaml:"requestTimeoutSeconds"`
}

// Environment represents a named upstream environment configuration.
type Environment struct {
	// Name is a unique identifier for this environment (e.g., "dev", "staging", "prod").
	Name string `yaml:"name"`

	// BaseURL is the upstream service URL. All RPCs to this environment will be proxied
	// to this base URL. This acts as an SSRF allowlist.
	BaseURL string `yaml:"baseURL"`

	// Transport specifies the default RPC transport for this environment.
	// Valid values: "connect", "grpc", "grpc-web".
	// Default: "connect".
	Transport string `yaml:"transport"`

	// TLS contains TLS-specific configuration for connecting to this environment.
	TLS TLSConfig `yaml:"tls"`

	// DefaultHeaders are headers that will be automatically included with every
	// request to this environment. Supports environment variable expansion.
	// Example: "x-api-key: ${REFLECT_DEV_API_KEY}"
	DefaultHeaders map[string]string `yaml:"defaultHeaders"`
}

// TLSConfig contains TLS-specific settings for an environment.
type TLSConfig struct {
	// InsecureSkipVerify disables certificate verification. Use only for development.
	// Default: false.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
}

// Default configuration values.
const (
	DefaultMaxRequestBodyBytes    = 1048576 // 1 MB
	DefaultRequestTimeoutSeconds  = 15
	DefaultTransport              = "connect"
)

// Load reads and parses a Reflect configuration file.
// It performs validation and applies default values.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	// Apply defaults
	if cfg.MaxRequestBodyBytes == 0 {
		cfg.MaxRequestBodyBytes = DefaultMaxRequestBodyBytes
	}
	if cfg.RequestTimeoutSeconds == 0 {
		cfg.RequestTimeoutSeconds = DefaultRequestTimeoutSeconds
	}

	// Expand environment variables in all config values
	if err := cfg.expandEnvVars(); err != nil {
		return nil, fmt.Errorf("expand environment variables: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars expands environment variables in all string fields of the config.
func (c *Config) expandEnvVars() error {
	for i := range c.Environments {
		env := &c.Environments[i]

		// Expand base URL
		env.BaseURL = os.Expand(env.BaseURL, os.Getenv)

		// Expand default headers
		for key, value := range env.DefaultHeaders {
			env.DefaultHeaders[key] = os.Expand(value, os.Getenv)
		}
	}
	return nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	// Check for duplicate environment names
	envNames := make(map[string]bool)
	for i := range c.Environments {
		env := &c.Environments[i]
		if envNames[env.Name] {
			return fmt.Errorf("duplicate environment name: %q", env.Name)
		}
		envNames[env.Name] = true

		// Validate environment
		if err := env.Validate(); err != nil {
			return fmt.Errorf("environment %q: %w", env.Name, err)
		}
	}

	// Validate limits
	if c.MaxRequestBodyBytes < 0 {
		return fmt.Errorf("maxRequestBodyBytes must be non-negative, got %d", c.MaxRequestBodyBytes)
	}
	if c.RequestTimeoutSeconds < 0 {
		return fmt.Errorf("requestTimeoutSeconds must be non-negative, got %d", c.RequestTimeoutSeconds)
	}

	return nil
}

// Validate checks that an environment configuration is valid.
func (e *Environment) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	if e.BaseURL == "" {
		return fmt.Errorf("baseURL is required")
	}

	// Validate base URL format
	parsedURL, err := url.Parse(e.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid baseURL: %w", err)
	}

	// Ensure base URL has a scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("baseURL must include a scheme (http:// or https://)")
	}

	// Ensure base URL has a host
	if parsedURL.Host == "" {
		return fmt.Errorf("baseURL must include a host")
	}

	// Validate transport if specified
	if e.Transport != "" {
		validTransports := map[string]bool{
			"connect":   true,
			"grpc":      true,
			"grpc-web":  true,
		}
		if !validTransports[e.Transport] {
			return fmt.Errorf("invalid transport %q, must be one of: connect, grpc, grpc-web", e.Transport)
		}
	} else {
		// Apply default transport
		e.Transport = DefaultTransport
	}

	return nil
}

// GetEnvironment retrieves an environment by name.
func (c *Config) GetEnvironment(name string) (*Environment, error) {
	for i := range c.Environments {
		if c.Environments[i].Name == name {
			return &c.Environments[i], nil
		}
	}
	return nil, fmt.Errorf("environment %q not found", name)
}

// IsHeaderAllowed checks if a header is in the allowlist.
// Header names are case-insensitive.
func (c *Config) IsHeaderAllowed(header string) bool {
	if len(c.HeaderAllowlist) == 0 {
		// If no allowlist is specified, allow all headers (permissive default)
		return true
	}

	headerLower := strings.ToLower(header)
	for _, allowed := range c.HeaderAllowlist {
		if strings.ToLower(allowed) == headerLower {
			return true
		}
	}
	return false
}

// GetTimeout returns the configured request timeout as a time.Duration.
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.RequestTimeoutSeconds) * time.Second
}
