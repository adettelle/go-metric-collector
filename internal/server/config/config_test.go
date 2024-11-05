package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromJSON(t *testing.T) {
	os.Setenv("CONFIG", "./testdata/test.cfg.json")
	defer os.Setenv("CONFIG", "")

	_, err := New(true)
	assert.NoError(t, err)
	// expectedCfg := Config{
	// 	Address:           "localhost:8080",
	// 	Key:               "",
	// 	CryptoKey:         "./keys/client_privatekey.pem",
	// 	ClientCert:        "./keys/client_cert.pem",
	// 	ServerCert:        "./keys/server_cert.pem",
	// 	Config:            "./testdata/test.cfg.json",
	// 	GrpcURL:           "",
	// 	MaxRequestRetries: 3,
	// 	PollInterval:      1,
	// 	ReportInterval:    10,
	// 	RateLimit:         1,
	// }
	// assert.Equal(t, cfg, &expectedCfg)
}

func TestDefaultConfig(t *testing.T) {
	cfg, err := New(false)
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

/*
func TestGetStoragePathInexistent(t *testing.T) {
	dir := os.TempDir()
	fileName := uuid.New().String()

	fullName := path.Join(dir, fileName)

	result := getStoragePath(&fullName)
	require.Equal(t, fullName, result)

	_, err := os.Stat(fullName)
	require.NoError(t, err)
}
*/
