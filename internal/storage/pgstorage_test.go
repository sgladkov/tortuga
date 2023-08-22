package storage

import (
	"testing"

	"database/sql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPgStorage_test(t *testing.T) {
	// This test requires local Postgres database to execute. Use `make postgresup' to start it in docker
	databaseDSN := "postgres://postgres:mysecretpassword@127.0.0.1:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", databaseDSN)
	require.NoError(t, err)
	testStorage, err = NewPgStorage(db)
	require.NoError(t, err)

	t.Run("Users", testUsers)
	t.Run("Projects", testProjects)
	t.Run("Bids", testBids)
	t.Run("Rates", testRates)
	t.Run("Transactions", testTransactions)
}
