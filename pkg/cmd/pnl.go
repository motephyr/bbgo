package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c9s/bbgo/pkg/accounting"
	"github.com/c9s/bbgo/pkg/accounting/pnl"
	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/service"
	"github.com/c9s/bbgo/pkg/types"
)

func init() {
	PnLCmd.Flags().String("session", "", "target exchange")
	PnLCmd.Flags().String("symbol", "", "trading symbol")
	PnLCmd.Flags().Int("limit", 500, "number of trades")
	RootCmd.AddCommand(PnLCmd)
}

var PnLCmd = &cobra.Command{
	Use:          "pnl",
	Short:        "pnl calculator",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		if len(configFile) == 0 {
			return errors.New("--config option is required")
		}

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return err
		}

		userConfig, err := bbgo.Load(configFile, false)
		if err != nil {
			return err
		}

		environ := bbgo.NewEnvironment()
		if err := environ.ConfigureDatabase(ctx) ; err != nil {
			return err
		}

		if err := environ.ConfigureExchangeSessions(userConfig); err != nil {
			return err
		}

		sessionName, err := cmd.Flags().GetString("session")
		if err != nil {
			return err
		}

		symbol, err := cmd.Flags().GetString("symbol")
		if err != nil {
			return err
		}
		if len(symbol) == 0 {
			return errors.New("--symbol [SYMBOL] is required")
		}

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return err
		}

		session, ok := environ.Session(sessionName)
		if !ok {
			return fmt.Errorf("session %s not found", sessionName)
		}

		exchange := session.Exchange

		var trades []types.Trade
		tradingFeeCurrency := exchange.PlatformFeeCurrency()
		if strings.HasPrefix(symbol, tradingFeeCurrency) {
			log.Infof("loading all trading fee currency related trades: %s", symbol)
			trades, err = environ.TradeService.QueryForTradingFeeCurrency(exchange.Name(), symbol, tradingFeeCurrency)
		} else {
			trades, err = environ.TradeService.Query(service.QueryTradesOptions{
				Exchange: exchange.Name(),
				Symbol:   symbol,
				Limit:	  limit,
			})
		}

		if err != nil {
			return err
		}

		log.Infof("%d trades loaded", len(trades))

		stockManager := &accounting.StockDistribution{
			Symbol:             symbol,
			TradingFeeCurrency: tradingFeeCurrency,
		}

		checkpoints, err := stockManager.AddTrades(trades)
		if err != nil {
			return err
		}

		log.Infof("found checkpoints: %+v", checkpoints)
		log.Infof("stock: %f", stockManager.Stocks.Quantity())

		tickers, err := exchange.QueryTickers(ctx, symbol)

		if err != nil {
			return err
		}

		currentTick, ok := tickers[symbol]

		if !ok {
			return errors.New("no ticker data for current price")
		}

		currentPrice := currentTick.Last

		calculator := &pnl.AverageCostCalculator{
			TradingFeeCurrency: tradingFeeCurrency,
		}

		report := calculator.Calculate(symbol, trades, currentPrice)
		report.Print()
		return nil
	},
}
