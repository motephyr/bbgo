package ftx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/c9s/bbgo/pkg/datatype"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

func TestExchange_QueryAccountBalances(t *testing.T) {
	successResp := `
{
  "result": [
    {
      "availableWithoutBorrow": 19.47458865,
      "coin": "USD",
      "free": 19.48085209,
      "spotBorrow": 0.0,
      "total": 1094.66405065,
      "usdValue": 1094.664050651561
    }
  ],
  "success": true
}
`
	failureResp := `{"result":[],"success":false}`
	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i == 0 {
			fmt.Fprintln(w, successResp)
			i++
			return
		}
		fmt.Fprintln(w, failureResp)
	}))
	defer ts.Close()

	ex := NewExchange("", "", "")
	serverURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	ex.restEndpoint = serverURL
	resp, err := ex.QueryAccountBalances(context.Background())
	assert.NoError(t, err)

	assert.Len(t, resp, 1)
	b, ok := resp["USD"]
	assert.True(t, ok)
	expectedAvailable := fixedpoint.Must(fixedpoint.NewFromString("19.48085209"))
	assert.Equal(t, expectedAvailable, b.Available)
	assert.Equal(t, fixedpoint.Must(fixedpoint.NewFromString("1094.66405065")).Sub(expectedAvailable), b.Locked)

	resp, err = ex.QueryAccountBalances(context.Background())
	assert.EqualError(t, err, "ftx returns querying balances failure")
	assert.Nil(t, resp)
}

func TestExchange_QueryOpenOrders(t *testing.T) {
	successResp := `
{
  "success": true,
  "result": [
    {
      "createdAt": "2019-03-05T09:56:55.728933+00:00",
      "filledSize": 10,
      "future": "XRP-PERP",
      "id": 9596912,
      "market": "XRP-PERP",
      "price": 0.306525,
      "avgFillPrice": 0.306526,
      "remainingSize": 31421,
      "side": "sell",
      "size": 31431,
      "status": "open",
      "type": "limit",
      "reduceOnly": false,
      "ioc": false,
      "postOnly": false,
      "clientId": null
    }
  ]
}
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, successResp)
	}))
	defer ts.Close()

	ex := NewExchange("", "", "")
	serverURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	ex.restEndpoint = serverURL
	resp, err := ex.QueryOpenOrders(context.Background(), "XRP-PREP")
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "XRP-PERP", resp[0].Symbol)
}

func TestExchange_QueryClosedOrders(t *testing.T) {
	t.Run("no closed orders", func(t *testing.T) {
		successResp := `{"success": true, "result": []}`

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, successResp)
		}))
		defer ts.Close()

		ex := NewExchange("", "", "")
		serverURL, err := url.Parse(ts.URL)
		assert.NoError(t, err)
		ex.restEndpoint = serverURL
		resp, err := ex.QueryClosedOrders(context.Background(), "BTC-PERP", time.Now(), time.Now(), 100)
		assert.NoError(t, err)

		assert.Len(t, resp, 0)
	})
	t.Run("one closed order", func(t *testing.T) {
		successResp := `
{
  "success": true,
  "result": [
    {
      "avgFillPrice": 10135.25,
      "clientId": null,
      "createdAt": "2019-06-27T15:24:03.101197+00:00",
      "filledSize": 0.001,
      "future": "BTC-PERP",
      "id": 257132591,
      "ioc": false,
      "market": "BTC-PERP",
      "postOnly": false,
      "price": 10135.25,
      "reduceOnly": false,
      "remainingSize": 0.0,
      "side": "buy",
      "size": 0.001,
      "status": "closed",
      "type": "limit"
    }
  ],
  "hasMoreData": false
}
`
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, successResp)
		}))
		defer ts.Close()

		ex := NewExchange("", "", "")
		serverURL, err := url.Parse(ts.URL)
		assert.NoError(t, err)
		ex.restEndpoint = serverURL
		resp, err := ex.QueryClosedOrders(context.Background(), "BTC-PERP", time.Now(), time.Now(), 100)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "BTC-PERP", resp[0].Symbol)
	})

	t.Run("sort the order", func(t *testing.T) {
		successResp := `
{
  "success": true,
  "result": [
    {
			"status": "closed",
      "createdAt": "2020-09-01T15:24:03.101197+00:00",
      "id": 789
    },
    {
			"status": "closed",
      "createdAt": "2019-03-27T15:24:03.101197+00:00",
      "id": 123
    },
    {
			"status": "closed",
      "createdAt": "2019-06-27T15:24:03.101197+00:00",
      "id": 456
    },
    {
			"status": "new",
      "createdAt": "2019-06-27T15:24:03.101197+00:00",
      "id": 999
    }
  ],
  "hasMoreData": false
}
`
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, successResp)
		}))
		defer ts.Close()

		ex := NewExchange("", "", "")
		serverURL, err := url.Parse(ts.URL)
		assert.NoError(t, err)
		ex.restEndpoint = serverURL
		resp, err := ex.QueryClosedOrders(context.Background(), "BTC-PERP", time.Now(), time.Now(), 100)
		assert.NoError(t, err)
		assert.Len(t, resp, 3)

		expectedOrderID := []uint64{123, 456, 789}
		for i, o := range resp {
			assert.Equal(t, expectedOrderID[i], o.OrderID)
		}
	})

	t.Run("receive duplicated orders", func(t *testing.T) {
		successRespOne := `
{
  "success": true,
  "result": [
    {
			"status": "closed",
      "createdAt": "2020-09-01T15:24:03.101197+00:00",
      "id": 123
    }
  ],
  "hasMoreData": true
}
`

		successRespTwo := `
{
  "success": true,
  "result": [
    {
			"clientId": "ignored-by-created-at",
			"status": "closed",
      "createdAt": "1999-09-01T15:24:03.101197+00:00",
      "id": 999
    },
    {
			"clientId": "ignored-by-duplicated-id",
			"status": "closed",
      "createdAt": "2020-09-02T15:24:03.101197+00:00",
      "id": 123
    },
    {
			"clientId": "ignored-duplicated-entry",
			"status": "closed",
      "createdAt": "2020-09-01T15:24:03.101197+00:00",
      "id": 123
    },
    {
			"status": "closed",
      "createdAt": "2020-09-02T15:24:03.101197+00:00",
      "id": 456
    }
  ],
  "hasMoreData": false
}
`
		i := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if i == 0 {
				i++
				fmt.Fprintln(w, successRespOne)
				return
			}
			fmt.Fprintln(w, successRespTwo)
		}))
		defer ts.Close()

		ex := NewExchange("", "", "")
		serverURL, err := url.Parse(ts.URL)
		assert.NoError(t, err)
		ex.restEndpoint = serverURL
		resp, err := ex.QueryClosedOrders(context.Background(), "BTC-PERP", time.Now(), time.Now(), 100)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		expectedOrderID := []uint64{123, 456}
		for i, o := range resp {
			assert.Equal(t, expectedOrderID[i], o.OrderID)
		}
	})
}

func TestExchange_QueryAccount(t *testing.T) {
	balanceResp := `
{
  "result": [
    {
      "availableWithoutBorrow": 19.47458865,
      "coin": "USD",
      "free": 19.48085209,
      "spotBorrow": 0.0,
      "total": 1094.66405065,
      "usdValue": 1094.664050651561
    }
  ],
  "success": true
}
`

	accountInfoResp := `
{
  "success": true,
  "result": {
    "backstopProvider": true,
    "collateral": 3568181.02691129,
    "freeCollateral": 1786071.456884368,
    "initialMarginRequirement": 0.12222384240257728,
    "leverage": 10,
    "liquidating": false,
    "maintenanceMarginRequirement": 0.07177992558058484,
    "makerFee": 0.0002,
    "marginFraction": 0.5588433331419503,
    "openMarginFraction": 0.2447194090423075,
    "takerFee": 0.0005,
    "totalAccountValue": 3568180.98341129,
    "totalPositionSize": 6384939.6992,
    "username": "user@domain.com",
    "positions": [
      {
        "cost": -31.7906,
        "entryPrice": 138.22,
        "future": "ETH-PERP",
        "initialMarginRequirement": 0.1,
        "longOrderSize": 1744.55,
        "maintenanceMarginRequirement": 0.04,
        "netSize": -0.23,
        "openSize": 1744.32,
        "realizedPnl": 3.39441714,
        "shortOrderSize": 1732.09,
        "side": "sell",
        "size": 0.23,
        "unrealizedPnl": 0
      }
    ]
  }
}
`
	returnBalance := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if returnBalance {
			fmt.Fprintln(w, balanceResp)
			return
		}
		returnBalance = true
		fmt.Fprintln(w, accountInfoResp)
	}))
	defer ts.Close()

	ex := NewExchange("", "", "")
	serverURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	ex.restEndpoint = serverURL
	resp, err := ex.QueryAccount(context.Background())
	assert.NoError(t, err)

	b, ok := resp.Balance("USD")
	assert.True(t, ok)
	expected := types.Balance{
		Currency:  "USD",
		Available: fixedpoint.MustNewFromString("19.48085209"),
		Locked:    fixedpoint.MustNewFromString("1094.66405065"),
	}
	expected.Locked = expected.Locked.Sub(expected.Available)
	assert.Equal(t, expected, b)

	assert.Equal(t, fixedpoint.NewFromFloat(0.0002), resp.MakerCommission)
	assert.Equal(t, fixedpoint.NewFromFloat(0.0005), resp.TakerCommission)
}

func TestExchange_QueryMarkets(t *testing.T) {
	respJSON := `{
"success": true,
"result": [
  {
    "name": "BTC/USD",
    "enabled": true,
    "postOnly": false,
    "priceIncrement": 1.0,
    "sizeIncrement": 0.0001,
    "minProvideSize": 0.001,
    "last": 59039.0,
    "bid": 59038.0,
    "ask": 59040.0,
    "price": 59039.0,
    "type": "spot",
    "baseCurrency": "BTC",
    "quoteCurrency": "USD",
    "underlying": null,
    "restricted": false,
    "highLeverageFeeExempt": true,
    "change1h": 0.0015777151969599294,
    "change24h": 0.05475756601279165,
    "changeBod": -0.0035107262814994852,
    "quoteVolume24h": 316493675.5463,
    "volumeUsd24h": 316493675.5463
  }
]
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, respJSON)
	}))
	defer ts.Close()

	ex := NewExchange("", "", "")
	serverURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	ex.restEndpoint = serverURL
	resp, err := ex.QueryMarkets(context.Background())
	assert.NoError(t, err)

	assert.Len(t, resp, 1)
	assert.Equal(t, types.Market{
		Symbol:          "BTC/USD",
		PricePrecision:  0,
		VolumePrecision: 4,
		QuoteCurrency:   "USD",
		BaseCurrency:    "BTC",
		MinQuantity:     0.001,
		StepSize:        0.0001,
		TickSize:        1,
	}, resp["BTC/USD"])
}

func TestExchange_QueryDepositHistory(t *testing.T) {
	respJSON := `
{
  "success": true,
  "result": [
    {
      "coin": "TUSD",
      "confirmations": 64,
      "confirmedTime": "2019-03-05T09:56:55.728933+00:00",
      "fee": 0,
      "id": 1,
      "sentTime": "2019-03-05T09:56:55.735929+00:00",
      "size": 99.0,
      "status": "confirmed",
      "time": "2019-03-05T09:56:55.728933+00:00",
      "txid": "0x8078356ae4b06a036d64747546c274af19581f1c78c510b60505798a7ffcaf1",
			"address": {"address": "test-addr", "tag": "test-tag"}
    }
  ]
}
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, respJSON)
	}))
	defer ts.Close()

	ex := NewExchange("", "", "")
	serverURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	ex.restEndpoint = serverURL

	ctx := context.Background()
	layout := "2006-01-02T15:04:05.999999Z07:00"
	actualConfirmedTime, err := time.Parse(layout, "2019-03-05T09:56:55.728933+00:00")
	assert.NoError(t, err)
	dh, err := ex.QueryDepositHistory(ctx, "TUSD", actualConfirmedTime.Add(-1*time.Hour), actualConfirmedTime.Add(1*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, dh, 1)
	assert.Equal(t, types.Deposit{
		Exchange:      types.ExchangeFTX,
		Time:          datatype.Time(actualConfirmedTime),
		Amount:        99.0,
		Asset:         "TUSD",
		TransactionID: "0x8078356ae4b06a036d64747546c274af19581f1c78c510b60505798a7ffcaf1",
		Status:        types.DepositSuccess,
		Address:       "test-addr",
		AddressTag:    "test-tag",
	}, dh[0])

	// not in the time range
	dh, err = ex.QueryDepositHistory(ctx, "TUSD", actualConfirmedTime.Add(1*time.Hour), actualConfirmedTime.Add(2*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, dh, 0)

	// exclude by asset
	dh, err = ex.QueryDepositHistory(ctx, "BTC", actualConfirmedTime.Add(-1*time.Hour), actualConfirmedTime.Add(1*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, dh, 0)
}
