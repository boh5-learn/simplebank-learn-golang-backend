package db_test

import (
	"database/sql"
	db "github.com/boh5-learn/simplebank-learn-golang-backend/db/sqlc"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *db.Queries

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testQueries = db.New(conn)

	os.Exit(m.Run())
}
