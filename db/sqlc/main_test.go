package db_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/boh5-learn/simplebank-learn-golang-backend/util"

	db "github.com/boh5-learn/simplebank-learn-golang-backend/db/sqlc"

	_ "github.com/lib/pq"
)

var (
	testQueries *db.Queries
	testDB      *sql.DB
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	// 这里不能用 := 语法，否则 testDB 是新初始化的局部变量，全局变量 testDB 是 nil
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testQueries = db.New(testDB)

	os.Exit(m.Run())
}
