package token_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/boh5-learn/simplebank-learn-golang-backend/token"
	"github.com/boh5-learn/simplebank-learn-golang-backend/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	t.Parallel()

	maker, err := token.NewJWTMaker(util.RandomString(32))
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

func TestExpiredJWTMaker(t *testing.T) {
	t.Parallel()

	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	tokenStr, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	payload, err := maker.VerifyToken(tokenStr)
	require.Error(t, err)
	require.EqualError(t, err, token.ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTMakerAlgNone(t *testing.T) {
	t.Parallel()

	payload, err := token.NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	tokenStr, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(tokenStr)
	require.Error(t, err)
	require.EqualError(t, err, token.ErrInvalidToken.Error())
	require.Nil(t, payload)
}
