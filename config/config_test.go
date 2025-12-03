package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfigYML_FileErrors(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		cfg, err := ReadConfigYML("/non/existent/path/config.yaml")

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("permission denied", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(filePath, []byte("test"), 0000)
		require.NoError(t, err)

		cfg, err := ReadConfigYML(filePath)

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestReadConfigYML_DecodeErrors(t *testing.T) {
	invalidYAML := `
project:
  name: "test"
  debug: true
invalid yaml content here
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(filePath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	cfg, err := ReadConfigYML(filePath)

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "yaml:")
}

func TestReadConfigYML_Success(t *testing.T) {
	yamlContent := `
project:
  name: "test-service"
  debug: true
  version: "1.0.0"
  commit_hash: "abc123"

opentelemetry:
  service: "opentelemetry-service"
  host: "localhost"
  port: 6831

status:
  livenessPath: "/healthz"
  readinessPath: "/readyz"
  versionPath: "/version"
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(filePath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := ReadConfigYML(filePath)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "test-service", cfg.Project.Name)
	assert.True(t, cfg.Project.Debug)
	assert.Equal(t, "/healthz", cfg.Status.LivenessPath)
	assert.Equal(t, "/readyz", cfg.Status.ReadinessPath)
	assert.Equal(t, "/version", cfg.Status.VersionPath)

	assert.Equal(t, version, cfg.Project.Version)
	assert.Equal(t, commitHash, cfg.Project.CommitHash)
}
