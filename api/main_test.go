package api_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/boh5-learn/simplebank-learn-golang-backend/util"

	"github.com/boh5-learn/simplebank-learn-golang-backend/api"
	db "github.com/boh5-learn/simplebank-learn-golang-backend/db/sqlc"

	"github.com/gin-gonic/gin"
)

func NewTestServer(t *testing.T, store db.Store) *api.Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := api.NewServer(config, store)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
