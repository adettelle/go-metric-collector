package migrator

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	dbParams := "host=localhost port=9999 user=postgres password=123456 dbname=test_db sslmode=disable"
	err := ApplyMigrations(dbParams)
	require.NoError(t, err)

	defer func() {
		if err := ResetMigrations(dbParams); err != nil {
			t.Fatal(err)
		}
	}()
}
