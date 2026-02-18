package config

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigFrom(t *testing.T) {
	t.Run("nonexistent file returns defaults", func(t *testing.T) {
		cfg, err := LoadConfigFrom("/nonexistent/path/signet.conf")
		if err != nil {
			t.Fatalf("LoadConfigFrom() error = %v", err)
		}

		if cfg.RootDir != defaultRootDir {
			t.Errorf("RootDir = %v, want %v", cfg.RootDir, defaultRootDir)
		}
		if cfg.Port != defaultPort {
			t.Errorf("Port = %v, want %v", cfg.Port, defaultPort)
		}
	})

	t.Run("existing file with values", func(t *testing.T) {
		tmpDir := t.TempDir()
		confPath := filepath.Join(tmpDir, "signet.conf")

		content := `RootDir = /custom/signet
Address = 10.0.0.1
NickName = TestUser
NodeName = testnode
Port = 9090
`
		if err := writeFile(confPath, content); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := LoadConfigFrom(confPath)
		if err != nil {
			t.Fatalf("LoadConfigFrom() error = %v", err)
		}

		if cfg.RootDir != "/custom/signet" {
			t.Errorf("RootDir = %v, want /custom/signet", cfg.RootDir)
		}
		if cfg.Address != "10.0.0.1" {
			t.Errorf("Address = %v, want 10.0.0.1", cfg.Address)
		}
		if cfg.NickName != "TestUser" {
			t.Errorf("NickName = %v, want TestUser", cfg.NickName)
		}
		if cfg.NodeName != "testnode" {
			t.Errorf("NodeName = %v, want testnode", cfg.NodeName)
		}
		if cfg.Port != "9090" {
			t.Errorf("Port = %v, want 9090", cfg.Port)
		}
	})

	t.Run("partial config uses defaults for missing values", func(t *testing.T) {
		tmpDir := t.TempDir()
		confPath := filepath.Join(tmpDir, "signet.conf")

		content := `Address = 10.0.0.1
NickName = TestUser
`
		if err := writeFile(confPath, content); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := LoadConfigFrom(confPath)
		if err != nil {
			t.Fatalf("LoadConfigFrom() error = %v", err)
		}

		if cfg.RootDir != defaultRootDir {
			t.Errorf("RootDir = %v, want %v", cfg.RootDir, defaultRootDir)
		}
		if cfg.Address != "10.0.0.1" {
			t.Errorf("Address = %v, want 10.0.0.1", cfg.Address)
		}
		if cfg.NickName != "TestUser" {
			t.Errorf("NickName = %v, want TestUser", cfg.NickName)
		}
		if cfg.Port != defaultPort {
			t.Errorf("Port = %v, want %v", cfg.Port, defaultPort)
		}
	})
}

func TestConfigPathHelpers(t *testing.T) {
	cfg := &Config{
		RootDir: "/test/signet",
	}

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{
			name:     "PrivKeyPath",
			method:   cfg.PrivKeyPath,
			expected: "/test/signet/ed25519.priv",
		},
		{
			name:     "BlockFilePath",
			method:   cfg.BlockFilePath,
			expected: "/test/signet/block.jsonl",
		},
		{
			name:     "PendingFilePath",
			method:   cfg.PendingFilePath,
			expected: "/test/signet/pending_transaction.json",
		},
		{
			name:     "NodesDir",
			method:   cfg.NodesDir,
			expected: "/test/signet/nodes",
		},
		{
			name:     "PIDFilePath",
			method:   cfg.PIDFilePath,
			expected: "/test/signet/signet.pid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestNodeFilePath(t *testing.T) {
	cfg := &Config{
		RootDir: "/test/signet",
	}

	expected := "/test/signet/nodes/node1"
	if got := cfg.NodeFilePath("node1"); got != expected {
		t.Errorf("NodeFilePath() = %v, want %v", got, expected)
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantHost  string
		wantPort  string
	}{
		{
			name:     "host with port",
			addr:     "10.0.0.1:8080",
			wantHost: "10.0.0.1",
			wantPort: "8080",
		},
		{
			name:     "host without port",
			addr:     "10.0.0.1",
			wantHost: "10.0.0.1",
			wantPort: defaultPort,
		},
		{
			name:     "localhost with port",
			addr:     "localhost:9090",
			wantHost: "localhost",
			wantPort: "9090",
		},
		{
			name:     "localhost without port",
			addr:     "localhost",
			wantHost: "localhost",
			wantPort: defaultPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := ParseAddress(tt.addr)
			if host != tt.wantHost {
				t.Errorf("ParseAddress() host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("ParseAddress() port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}
