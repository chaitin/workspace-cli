package safelinece

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("load from file", func(t *testing.T) {
		// 创建临时配置文件
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `safeline-ce:
  url: "https://example.com:9443"
  api_key: "test-token-123"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if cfg.URL != "https://example.com:9443" {
			t.Errorf("URL = %q, want %q", cfg.URL, "https://example.com:9443")
		}
		if cfg.APIKey != "test-token-123" {
			t.Errorf("APIKey = %q, want %q", cfg.APIKey, "test-token-123")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("LoadConfig() expected error for nonexistent file")
		}
	})

	t.Run("missing safeline-ce section", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `other-product:
  url: "https://example.com"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() expected error for missing safeline-ce section")
		}
	})

	t.Run("missing url", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `safeline-ce:
  api_key: "test-token"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() expected error for missing url")
		}
	})

	t.Run("missing api_key", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `safeline-ce:
  url: "https://example.com"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() expected error for missing api_key")
		}
	})
}

func TestConfig_ApplyEnvOverrides(t *testing.T) {
	cfg := &Config{
		URL:    "https://original.com",
		APIKey: "original-token",
	}

	// 设置环境变量
	os.Setenv("SAFELINE_CE_URL", "https://override.com")
	os.Setenv("SAFELINE_CE_API_KEY", "override-token")
	defer func() {
		os.Unsetenv("SAFELINE_CE_URL")
		os.Unsetenv("SAFELINE_CE_API_KEY")
	}()

	cfg.ApplyEnvOverrides()

	if cfg.URL != "https://override.com" {
		t.Errorf("URL = %q, want %q", cfg.URL, "https://override.com")
	}
	if cfg.APIKey != "override-token" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "override-token")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				URL:    "https://example.com",
				APIKey: "token",
			},
			wantErr: false,
		},
		{
			name: "missing url",
			config: Config{
				APIKey: "token",
			},
			wantErr: true,
		},
		{
			name: "missing api_key",
			config: Config{
				URL: "https://example.com",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	// 验证退出码常量值
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitAPIError != 1 {
		t.Errorf("ExitAPIError = %d, want 1", ExitAPIError)
	}
	if ExitNetworkError != 2 {
		t.Errorf("ExitNetworkError = %d, want 2", ExitNetworkError)
	}
	if ExitConfigError != 3 {
		t.Errorf("ExitConfigError = %d, want 3", ExitConfigError)
	}
}
