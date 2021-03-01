package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/c9s/bbgo/pkg/types"
)

type TransferService struct {
	DB *sqlx.DB
}

func (s *TransferService) InsertDeposit(deposit types.Deposit) error {
	sql := `INSERT INTO deposits (exchange, asset, address, amount, txn_id, time)
			VALUES (:exchange, :asset, :address, :amount, :txn_id, :time)`
	_, err := s.DB.NamedExec(sql, deposit)
	return err
}

func (s *TransferService) InsertWithdrawal(withdrawal types.Withdrawal) error {
	sql := `INSERT INTO withdraws (exchange, asset, network, address, amount, txn_id, txn_fee, time)
			VALUES (:exchange, :asset, :network, :address, :amount, :txn_id, :txn_fee, :time)`
	_, err := s.DB.NamedExec(sql, withdrawal)
	return err
}

