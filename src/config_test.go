package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"INPUT":  "/tmp/book",
				"OUTPUT": "./test.epub",
				"COVER":  "./assets/cover.jpg",
				"STYLE":  "./assets/style.css",
				"UUID":   "12345678-1234-1234-1234-123456789012",
			},
			wantError: false,
		},
		{
			name: "missing uuid",
			envVars: map[string]string{
				"INPUT":  "/tmp/book",
				"OUTPUT": "./test.epub",
				"COVER":  "./assets/cover.jpg",
				"STYLE":  "./assets/style.css",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := LoadConfig()

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if cfg == nil {
					t.Error("expected config but got nil")
				}
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	// Set only required UUID
	os.Setenv("UUID", "12345678-1234-1234-1234-123456789012")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check default values
	if cfg.Input != "/tmp/book" {
		t.Errorf("expected default Input '/tmp/book', got '%s'", cfg.Input)
	}
	if cfg.Output != "./il-manuale-del-buon-dev.epub" {
		t.Errorf("expected default Output './il-manuale-del-buon-dev.epub', got '%s'", cfg.Output)
	}
	if cfg.Cover != "./assets/cover.jpg" {
		t.Errorf("expected default Cover './assets/cover.jpg', got '%s'", cfg.Cover)
	}
	if cfg.Style != "./assets/style.css" {
		t.Errorf("expected default Style './assets/style.css', got '%s'", cfg.Style)
	}
}
