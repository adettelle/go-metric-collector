package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInterrupt(t *testing.T) {
	os.Setenv("SERVER_CERT", "./testdata/cert.pem")
	os.Setenv("CLIENT_CERT", "./testdata/client_cert.pem")
	os.Setenv("CRYPTO_KEY", "./testdata/client_privatekey.pem")
	defer func() {
		os.Setenv("SERVER_CERT", "")
		os.Setenv("CLIENT_CERT", "")
		os.Setenv("CRYPTO_KEY", "")
	}()

	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	err := initialize()
	require.NoError(t, err)
}
