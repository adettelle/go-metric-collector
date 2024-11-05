package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInterrupt(t *testing.T) {
	os.Setenv("CERT", "./testdata/cert.pem")
	os.Setenv("ADDRESS", ":12121")
	os.Setenv("CRYPTO_KEY", "./testdata/server_privatekey.pem")
	defer func() {
		os.Setenv("CERT", "")
		os.Setenv("CRYPTO_KEY", "")
	}()

	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	err := initialize()
	require.NoError(t, err)
}
