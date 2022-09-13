package util_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/boh5-learn/simplebank-learn-golang-backend/util"
	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	t.Parallel()

	password := util.RandomString(6)
	hashedPassword1, err := util.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = util.CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	wrongPassword := util.RandomString(7)
	err = util.CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPassword2, err := util.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
