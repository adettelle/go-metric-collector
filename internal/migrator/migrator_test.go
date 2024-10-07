package migrator

import (
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(9999).Database("metrics-test"))

	if err := database.Start(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	err := ApplyMigrations("host=localhost port=9999 user=postgres password=postgres dbname=metrics-test sslmode=disable")
	require.NoError(t, err)
}
