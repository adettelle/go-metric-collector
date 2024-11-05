package mware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	f := WithLogging(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	writer := httptest.NewRecorder()

	f(writer, httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")))
	require.Equal(t, writer.Code, 200)
}
