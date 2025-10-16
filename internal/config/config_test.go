package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		yamlConfig string
		envVars    map[string]string
		wantErr    bool
		validate   func(*testing.T, *Config)
	}{
		{
			name: "valid config with all fields",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev.api.example.com
    transport: connect
    tls:
      insecureSkipVerify: false
    defaultHeaders:
      x-api-key: secret123
  - name: prod
    baseURL: https://api.example.com
    transport: grpc
headerAllowlist:
  - authorization
  - x-api-key
maxRequestBodyBytes: 2097152
requestTimeoutSeconds: 30
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Environments) != 2 {
					t.Errorf("expected 2 environments, got %d", len(cfg.Environments))
				}
				if cfg.Environments[0].Name != "dev" {
					t.Errorf("expected first env name 'dev', got %q", cfg.Environments[0].Name)
				}
				if cfg.Environments[0].Transport != "connect" {
					t.Errorf("expected transport 'connect', got %q", cfg.Environments[0].Transport)
				}
				if cfg.MaxRequestBodyBytes != 2097152 {
					t.Errorf("expected maxRequestBodyBytes 2097152, got %d", cfg.MaxRequestBodyBytes)
				}
				if cfg.RequestTimeoutSeconds != 30 {
					t.Errorf("expected requestTimeoutSeconds 30, got %d", cfg.RequestTimeoutSeconds)
				}
				if len(cfg.HeaderAllowlist) != 2 {
					t.Errorf("expected 2 allowed headers, got %d", len(cfg.HeaderAllowlist))
				}
			},
		},
		{
			name: "minimal config with defaults",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev.api.example.com
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.MaxRequestBodyBytes != DefaultMaxRequestBodyBytes {
					t.Errorf("expected default maxRequestBodyBytes %d, got %d", DefaultMaxRequestBodyBytes, cfg.MaxRequestBodyBytes)
				}
				if cfg.RequestTimeoutSeconds != DefaultRequestTimeoutSeconds {
					t.Errorf("expected default requestTimeoutSeconds %d, got %d", DefaultRequestTimeoutSeconds, cfg.RequestTimeoutSeconds)
				}
				if cfg.Environments[0].Transport != DefaultTransport {
					t.Errorf("expected default transport %q, got %q", DefaultTransport, cfg.Environments[0].Transport)
				}
			},
		},
		{
			name: "environment variable expansion",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://${TEST_HOST}
    defaultHeaders:
      x-api-key: ${TEST_API_KEY}
      x-static: static-value
`,
			envVars: map[string]string{
				"TEST_HOST":    "dev.example.com",
				"TEST_API_KEY": "secret123",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Environments[0].BaseURL != "https://dev.example.com" {
					t.Errorf("expected baseURL with expanded var, got %q", cfg.Environments[0].BaseURL)
				}
				if cfg.Environments[0].DefaultHeaders["x-api-key"] != "secret123" {
					t.Errorf("expected expanded header value, got %q", cfg.Environments[0].DefaultHeaders["x-api-key"])
				}
				if cfg.Environments[0].DefaultHeaders["x-static"] != "static-value" {
					t.Errorf("expected static value, got %q", cfg.Environments[0].DefaultHeaders["x-static"])
				}
			},
		},
		{
			name: "duplicate environment names",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev1.api.example.com
  - name: dev
    baseURL: https://dev2.api.example.com
`,
			wantErr: true,
		},
		{
			name: "invalid base URL - no scheme",
			yamlConfig: `
environments:
  - name: dev
    baseURL: dev.api.example.com
`,
			wantErr: true,
		},
		{
			name: "invalid base URL - no host",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://
`,
			wantErr: true,
		},
		{
			name: "invalid transport",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev.api.example.com
    transport: invalid
`,
			wantErr: true,
		},
		{
			name: "missing environment name",
			yamlConfig: `
environments:
  - baseURL: https://dev.api.example.com
`,
			wantErr: true,
		},
		{
			name: "missing base URL",
			yamlConfig: `
environments:
  - name: dev
`,
			wantErr: true,
		},
		{
			name: "negative max request body bytes",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev.api.example.com
maxRequestBodyBytes: -1
`,
			wantErr: true,
		},
		{
			name: "negative timeout",
			yamlConfig: `
environments:
  - name: dev
    baseURL: https://dev.api.example.com
requestTimeoutSeconds: -1
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for this test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "reflect.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yamlConfig), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Load config
			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Run validation if provided
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/reflect.yaml")
	if err == nil {
		t.Error("expected error loading nonexistent file, got nil")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reflect.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error parsing invalid YAML, got nil")
	}
}

func TestGetEnvironment(t *testing.T) {
	cfg := &Config{
		Environments: []Environment{
			{Name: "dev", BaseURL: "https://dev.example.com"},
			{Name: "prod", BaseURL: "https://api.example.com"},
		},
	}

	tests := []struct {
		name    string
		envName string
		wantErr bool
		wantURL string
	}{
		{
			name:    "existing environment",
			envName: "dev",
			wantErr: false,
			wantURL: "https://dev.example.com",
		},
		{
			name:    "nonexistent environment",
			envName: "staging",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := cfg.GetEnvironment(tt.envName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && env.BaseURL != tt.wantURL {
				t.Errorf("expected baseURL %q, got %q", tt.wantURL, env.BaseURL)
			}
		})
	}
}

func TestIsHeaderAllowed(t *testing.T) {
	tests := []struct {
		name      string
		allowlist []string
		header    string
		want      bool
	}{
		{
			name:      "allowed header - exact match",
			allowlist: []string{"authorization", "x-api-key"},
			header:    "authorization",
			want:      true,
		},
		{
			name:      "allowed header - case insensitive",
			allowlist: []string{"authorization", "x-api-key"},
			header:    "Authorization",
			want:      true,
		},
		{
			name:      "disallowed header",
			allowlist: []string{"authorization", "x-api-key"},
			header:    "cookie",
			want:      false,
		},
		{
			name:      "empty allowlist - all allowed",
			allowlist: []string{},
			header:    "any-header",
			want:      true,
		},
		{
			name:      "nil allowlist - all allowed",
			allowlist: nil,
			header:    "any-header",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{HeaderAllowlist: tt.allowlist}
			got := cfg.IsHeaderAllowed(tt.header)
			if got != tt.want {
				t.Errorf("IsHeaderAllowed(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	cfg := &Config{RequestTimeoutSeconds: 30}
	expected := 30 * time.Second
	got := cfg.GetTimeout()
	if got != expected {
		t.Errorf("GetTimeout() = %v, want %v", got, expected)
	}
}

func TestEnvironmentValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     Environment
		wantErr bool
	}{
		{
			name: "valid environment",
			env: Environment{
				Name:      "dev",
				BaseURL:   "https://dev.example.com",
				Transport: "connect",
			},
			wantErr: false,
		},
		{
			name: "valid with grpc transport",
			env: Environment{
				Name:      "prod",
				BaseURL:   "https://api.example.com",
				Transport: "grpc",
			},
			wantErr: false,
		},
		{
			name: "valid with grpc-web transport",
			env: Environment{
				Name:      "prod",
				BaseURL:   "https://api.example.com",
				Transport: "grpc-web",
			},
			wantErr: false,
		},
		{
			name: "valid with http scheme",
			env: Environment{
				Name:    "local",
				BaseURL: "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			env: Environment{
				BaseURL: "https://api.example.com",
			},
			wantErr: true,
		},
		{
			name: "missing baseURL",
			env: Environment{
				Name: "dev",
			},
			wantErr: true,
		},
		{
			name: "invalid transport",
			env: Environment{
				Name:      "dev",
				BaseURL:   "https://api.example.com",
				Transport: "http",
			},
			wantErr: true,
		},
		{
			name: "malformed URL",
			env: Environment{
				Name:    "dev",
				BaseURL: "://invalid",
			},
			wantErr: true,
		},
		{
			name: "URL without scheme",
			env: Environment{
				Name:    "dev",
				BaseURL: "api.example.com",
			},
			wantErr: true,
		},
		{
			name: "URL without host",
			env: Environment{
				Name:    "dev",
				BaseURL: "https://",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			// If no transport was specified and validation passed, check default was applied
			if !tt.wantErr && tt.env.Transport == "" {
				t.Errorf("expected default transport to be applied, got empty string")
			}
		})
	}
}

func TestEnvironmentValidateAppliesDefaults(t *testing.T) {
	env := Environment{
		Name:    "dev",
		BaseURL: "https://dev.example.com",
		// Transport not specified
	}

	err := env.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if env.Transport != DefaultTransport {
		t.Errorf("expected default transport %q, got %q", DefaultTransport, env.Transport)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				Environments: []Environment{
					{Name: "dev", BaseURL: "https://dev.example.com", Transport: "connect"},
				},
				MaxRequestBodyBytes:   1048576,
				RequestTimeoutSeconds: 15,
			},
			wantErr: false,
		},
		{
			name: "duplicate environment names",
			cfg: Config{
				Environments: []Environment{
					{Name: "dev", BaseURL: "https://dev1.example.com", Transport: "connect"},
					{Name: "dev", BaseURL: "https://dev2.example.com", Transport: "connect"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate environment name",
		},
		{
			name: "invalid environment",
			cfg: Config{
				Environments: []Environment{
					{Name: "dev", BaseURL: "", Transport: "connect"},
				},
			},
			wantErr: true,
		},
		{
			name: "negative max request body bytes",
			cfg: Config{
				Environments: []Environment{
					{Name: "dev", BaseURL: "https://dev.example.com", Transport: "connect"},
				},
				MaxRequestBodyBytes: -100,
			},
			wantErr: true,
			errMsg:  "maxRequestBodyBytes must be non-negative",
		},
		{
			name: "negative timeout",
			cfg: Config{
				Environments: []Environment{
					{Name: "dev", BaseURL: "https://dev.example.com", Transport: "connect"},
				},
				RequestTimeoutSeconds: -5,
			},
			wantErr: true,
			errMsg:  "requestTimeoutSeconds must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.errMsg)) {
				t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}
