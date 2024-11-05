package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromJSON(t *testing.T) {
	cfg, err := New(true, "./testdata/test.cfg.json")
	assert.NoError(t, err)
	expectedCfg := Config{
		Address:       "localhost:8080",
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
		Config:        "./testdata/test.cfg.json",
		StoragePath:   "/tmp/metrics-db.json",
		CryptoKey:     "./keys/server_privatekey.pem",
		Cert:          "",
		TrustedSubnet: "",
		GrpcPort:      "",
		StoreInterval: 1,
		Restore:       true,
	}
	assert.Equal(t, cfg, &expectedCfg)
}

func TestDefaultConfig(t *testing.T) {
	cfg, err := New(false, "")
	require.Nil(t, err)

	require.Equal(t, &Config{
		Address:       "localhost:8080",
		StoreInterval: 300,
		StoragePath:   "/tmp/metrics-db.json",
		Restore:       true,
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
		GrpcPort:      "3200",
	}, cfg)
}

func TestShouldRestoreFileNotExists(t *testing.T) {
	cfg := &Config{
		Address:       "localhost:8080",
		StoreInterval: 300,
		StoragePath:   "/tmp/inexistent.json",
		Restore:       true,
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
	}

	require.False(t, cfg.ShouldRestore())
}

func TestShouldRestoreEmptyFileExists(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "testfile.json") // "/tmp",
	require.NoError(t, err)

	cfg := &Config{
		Address:       "localhost:8080",
		StoreInterval: 300,
		StoragePath:   file.Name(),
		Restore:       true,
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
	}

	require.False(t, cfg.ShouldRestore())
}

func TestShouldRestoreFileExists(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "testfile.json")
	require.NoError(t, err)
	err = os.WriteFile(file.Name(), []byte("{}"), 0700)
	require.NoError(t, err)

	cfg := &Config{
		Address:       "localhost:8080",
		StoreInterval: 300,
		StoragePath:   file.Name(),
		Restore:       true,
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
	}

	require.True(t, cfg.ShouldRestore())
}
