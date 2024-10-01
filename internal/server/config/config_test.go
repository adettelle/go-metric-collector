package config

import (
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// func TestGetKeyWithNoEnv(t *testing.T) {
// 	expected := "hello"
// 	res := getKey(&expected)
// 	require.Equal(t, res, expected)
// }

// func TestGetKeyWithEnv(t *testing.T) {
// 	param := "hello"
// 	os.Setenv("KEY", "keyFromEnv")
// 	res := getKey(&param)
// 	require.Equal(t, res, os.Getenv("KEY"))
// }

func TestDefaultConfig(t *testing.T) {
	cfg, err := New()
	require.Nil(t, err)

	require.Equal(t, &Config{
		Address:       "localhost:8080",
		StoreInterval: 300,
		StoragePath:   "/tmp/metrics-db.json",
		Restore:       true,
		DBParams:      "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		Key:           "",
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
	file, err := os.CreateTemp("/tmp", "testfile.json")
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
	file, err := os.CreateTemp("/tmp", "testfile.json")
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

func TestGetStoragePathInexistent(t *testing.T) {
	dir := os.TempDir()
	fileName := uuid.New().String()

	fullName := path.Join(dir, fileName)

	result := getStoragePath(&fullName)
	require.Equal(t, fullName, result)

	_, err := os.Stat(fullName)
	require.NoError(t, err)
}
