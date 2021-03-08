package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/c9s/bbgo/pkg/types"
)

type WithdrawService struct {
	DB *sqlx.DB
}

func (s *WithdrawService) Query(exchangeName types.ExchangeName) ([]types.Withdrawal, error) {
	args := map[string]interface{}{
		"exchange": exchangeName,
	}
	sql := "SELECT * FROM `withdraws` WHERE `exchange` = :exchange ORDER BY `time` ASC"
	rows, err := s.DB.NamedQuery(sql, args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *WithdrawService) scanRows(rows *sqlx.Rows) (withdraws []types.Withdrawal, err error) {
	for rows.Next() {
		var withdraw types.Withdrawal
		if err := rows.StructScan(&withdraw); err != nil {
			return withdraws, err
		}

		withdraws = append(withdraws, withdraw)
	}

	return withdraws, rows.Err()
}

func (s *WithdrawService) Insert(withdrawal types.Withdrawal) error {
	sql := `INSERT INTO withdraws (exchange, asset, network, address, amount, txn_id, txn_fee, time)
			VALUES (:exchange, :asset, :network, :address, :amount, :txn_id, :txn_fee, :time)`
	_, err := s.DB.NamedExec(sql, withdrawal)
	return err
}

