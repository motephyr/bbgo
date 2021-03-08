package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/c9s/bbgo/pkg/types"
)

type WithdrawService struct {
	DB *sqlx.DB
}

func (s *WithdrawService) InsertDeposit(deposit types.Deposit) error {
	sql := `INSERT INTO deposits (exchange, asset, address, amount, txn_id, time)
			VALUES (:exchange, :asset, :address, :amount, :txn_id, :time)`
	_, err := s.DB.NamedExec(sql, deposit)
	return err
}

func (s *WithdrawService) Insert(withdrawal types.Withdrawal) error {
	sql := `INSERT INTO withdraws (exchange, asset, network, address, amount, txn_id, txn_fee, time)
			VALUES (:exchange, :asset, :network, :address, :amount, :txn_id, :txn_fee, :time)`
	_, err := s.DB.NamedExec(sql, withdrawal)
	return err
}

