package pnl

import (
	"context"
	"time"

	"github.com/c9s/bbgo/pkg/datatype"
	"github.com/c9s/bbgo/pkg/exchange/batch"
	"github.com/c9s/bbgo/pkg/types"
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

type DailyAveragePriceMap map[Date]float64

func QueryExchangeDailyAveragePrices(ctx context.Context, ex types.Exchange, symbol string, since, until time.Time) (DailyAveragePriceMap, error) {
	q := batch.KLineBatchQuery{Exchange: ex}
	klineC, errC := q.Query(ctx, symbol, types.Interval1d, since, until)

	prices := make(DailyAveragePriceMap)
	for k := range klineC {
		prices[Date{
			Year:  k.StartTime.Year(),
			Month: k.StartTime.Month(),
			Day:   k.StartTime.Day(),
		}] = k.Close
	}

	if err := <-errC; err != nil {
		return nil, err
	}

	return prices, nil
}

// TransferConverter converts withdrawal records and deposit records into trades
// So that we can get the correct position of a PnL
type TransferConverter struct {
	BaseAsset  string
	QuoteAsset string

	dailyAveragePrices DailyAveragePriceMap
}

func (c *TransferConverter) convertWithdraws(withdraws []types.Withdraw) (trades []types.Trade, err error) {
	for _, withdrawal := range withdraws {
		trades = append(trades, types.Trade{
			GID:           0,
			ID:            0,
			OrderID:       0,
			Exchange:      "",
			Price:         0,
			Quantity:      0,
			QuoteQuantity: 0,
			Symbol:        c.BaseAsset + c.QuoteAsset,
			Side:          types.SideTypeSell,
			Time:          datatype.Time(withdrawal.EffectiveTime()),
			Fee:           withdrawal.TransactionFee,
			FeeCurrency:   withdrawal.Asset,
			IsMargin:      false,
			IsIsolated:    false,
		})
	}

	return
}

func (c *TransferConverter) QueryAndConvert(ctx context.Context, ex types.ExchangeTransferService, since, until time.Time) ([]types.Trade, error) {
	withdrawals, err := ex.QueryWithdrawHistory(ctx, c.BaseAsset, since, until)
	if err != nil {
		return nil, err
	}
	_ = withdrawals

	deposits, err := ex.QueryDepositHistory(ctx, c.BaseAsset, since, until)
	if err != nil {
		return nil, err
	}
	_ = deposits

	return nil, nil
}
