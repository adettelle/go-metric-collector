package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFromJSON(t *testing.T) {
	os.Setenv("CONFIG", "./testdata/test.cfg.json")
	defer os.Setenv("CONFIG", "")

	cfg, err := New()
	assert.NoError(t, err)
	expectedCfg := Config{
		Address:           "localhost:8080",
		Key:               "",
		CryptoKey:         "./keys/client_privatekey.pem",
		ClientCert:        "./keys/client_cert.pem",
		ServerCert:        "./keys/server_cert.pem",
		Config:            "./testdata/test.cfg.json",
		GrpcURL:           "",
		MaxRequestRetries: 3,
		PollInterval:      1,
		ReportInterval:    10,
		RateLimit:         1,
	}
	assert.Equal(t, cfg, &expectedCfg)
}
