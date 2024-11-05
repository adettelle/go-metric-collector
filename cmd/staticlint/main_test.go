package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetChecks(t *testing.T) {
	mychecks := getChecks()
	require.NotEmpty(t, mychecks)
}
