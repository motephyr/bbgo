package service

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/c9s/bbgo/pkg/datatype"
	"github.com/c9s/bbgo/pkg/types"
)

func TestTransferService(t *testing.T) {
	db, err := prepareDB(t)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	xdb := sqlx.NewDb(db.DB, "sqlite3")
	service := &TransferService{DB: xdb}

	err = service.InsertWithdrawal(types.Withdrawal{
		Exchange:       types.ExchangeMax,
		Asset:          "BTC",
		Amount:         0.0001,
		Address:        "test",
		TransactionID:  "01",
		TransactionFee: 0.0001,
		Network:        "omni",
		ApplyTime:      datatype.Time(time.Now()),
	})
	assert.NoError(t, err)

	err = service.InsertDeposit(types.Deposit{
		Exchange:      types.ExchangeMax,
		Time:          datatype.Time(time.Now()),
		Amount:        0.001,
		Asset:         "BTC",
		Address:       "test",
		TransactionID: "02",
		Status:        types.DepositSuccess,
	})
	assert.NoError(t, err)
}
