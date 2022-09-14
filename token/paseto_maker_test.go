package token_test

import (
	"testing"
	"time"

	"github.com/boh5-learn/simplebank-learn-golang-backend/token"
	"github.com/boh5-learn/simplebank-learn-golang-backend/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	t.Parallel()

	maker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	tokenStr, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	payload, err := maker.VerifyToken(tokenStr)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoMaker(t *testing.T) {
	t.Parallel()

	maker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	tokenStr, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	payload, err := maker.VerifyToken(tokenStr)
	require.Error(t, err)
	require.EqualError(t, err, token.ErrExpiredToken.Error())
	require.Nil(t, payload)
}
