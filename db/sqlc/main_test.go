package db_test

import (
	"database/sql"
	db "github.com/boh5-learn/simplebank-learn-golang-backend/db/sqlc"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	testQueries *db.Queries
	testDB      *sql.DB
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func TestMain(m *testing.M) {
	var err error

	// 这里不能用 := 语法，否则 testDB 是新初始化的局部变量，全局变量 testDB 是 nil
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testQueries = db.New(testDB)

	os.Exit(m.Run())
}
