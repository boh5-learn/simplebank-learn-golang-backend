package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boh5-learn/simplebank-learn-golang-backend/api"

	db "github.com/boh5-learn/simplebank-learn-golang-backend/db/sqlc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	mockdb "github.com/boh5-learn/simplebank-learn-golang-backend/db/mock"
	"github.com/boh5-learn/simplebank-learn-golang-backend/util"
	"github.com/gin-gonic/gin"
)

func TestCreateTransfer(t *testing.T) {
	t.Parallel()

	accountUSD1 := randomAccount()
	accountUSD2 := randomAccount()
	accountEUR1 := randomAccount()

	accountUSD1.Currency = util.USD
	accountUSD2.Currency = util.USD
	accountEUR1.Currency = util.EUR

	amount := int64(10)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD1.ID)).
					Times(1).
					Return(accountUSD1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD2.ID)).
					Times(1).
					Return(accountUSD2, nil)

				arg := db.TransferTxParams{
					FromAccountID: accountUSD1.ID,
					ToAccountID:   accountUSD2.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), arg).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD1.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD1.ID)).
					Times(1).
					Return(accountUSD1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD2.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": accountEUR1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountEUR1.ID)).
					Times(1).
					Return(accountEUR1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": accountEUR1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.EUR,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountEUR1.ID)).
					Times(1).
					Return(accountEUR1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD2.ID)).
					Times(1).
					Return(accountUSD2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        "INVALID_CURRENCY",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          -amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": accountUSD1.ID,
				"to_account_id":   accountUSD2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD1.ID)).
					Times(1).
					Return(accountUSD1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accountUSD2.ID)).
					Times(1).
					Return(accountUSD2, nil)

				arg := db.TransferTxParams{
					FromAccountID: accountUSD1.ID,
					ToAccountID:   accountUSD2.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), arg).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := api.NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router().ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
