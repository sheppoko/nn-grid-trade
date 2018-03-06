package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"nn-grid-trade/config"
	"nn-grid-trade/util"
	"strconv"
	"time"

	"github.com/go-resty/resty"
	"github.com/google/go-querystring/query"
)

const (
	PrivateAPIEndpoint    = "https://api.bitbank.cc"
	CandleAPIEndpointBase = "https://public.bitbank.cc/" + config.CoinName + "_" + config.CoinPairName + "/candlestick/1min/"
	BoardAPIEndpoint      = "https://public.bitbank.cc/" + config.CoinName + "_" + config.CoinPairName + "/depth"
	APIKey                = "7439d369-442c-4dd5-bc7d-01eed90a9bc1"
	AssetsPath            = "/v1/user/assets"
	ActiveOrdersPath      = "/v1/user/spot/active_orders"
	TradeHistoryPath      = "/v1/user/spot/trade_history"
	OrderPath             = "/v1/user/spot/order"
	CancelOrdersPath      = "/v1/user/spot/cancel_orders"
	OrdersInfoPath        = "/v1/user/spot/orders_info"
	SlackWebHookEndpoint  = "https://hooks.slack.com/services/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"
)

type Sec struct {
	Value string `json:"sec"`
}

var (
	apiSecret = &Sec{}
)

var nonceIncrementalCounter = 0.0

type BoardResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code      int        `json:"code"`
		Asks      [][]string `json:"asks"`
		Bids      [][]string `json:"bids"`
		Timestamp int64      `json:"timestamp"`
	} `json:"data"`
}

type AssetsResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code   int `json:"code"`
		Assets []struct {
			Asset           string  `json:"asset"`
			AmountPrecision int     `json:"amount_precision"`
			OnhandAmount    float64 `json:"onhand_amount,string"`
			LockedAmount    float64 `json:"locked_amount,string"`
			FreeAmount      float64 `json:"free_amount,string"`
			StopDeposit     bool    `json:"stop_deposit"`
			StopWithdrawal  bool    `json:"stop_withdrawal"`
		} `json:"assets"`
	} `json:"data"`
}

type ActiveOrdersResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code   int `json:"code"`
		Orders []struct {
			OrderID         int     `json:"order_id"`
			Pair            string  `json:"pair"`
			Side            string  `json:"side"`
			Type            string  `json:"type"`
			StartAmount     string  `json:"start_amount"`
			RemainingAmount float64 `json:"remaining_amount,string"`
			ExecutedAmount  float64 `json:"executed_amount,string"`
			Price           float64 `json:"price,string"`
			AveragePrice    string  `json:"average_price"`
			OrderedAt       int     `json:"ordered_at"`
			Status          string  `json:"status"`
		} `json:"orders"`
	} `json:"data"`
}

type OrdersInfoResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code   int `json:"code"`
		Orders []struct {
			OrderID         int     `json:"order_id"`
			Pair            string  `json:"pair"`
			Side            string  `json:"side"`
			Type            string  `json:"type"`
			StartAmount     string  `json:"start_amount"`
			RemainingAmount float64 `json:"remaining_amount,string"`
			ExecutedAmount  float64 `json:"executed_amount,string"`
			Price           float64 `json:"price,string"`
			AveragePrice    string  `json:"average_price"`
			OrderedAt       int     `json:"ordered_at"`
			Status          string  `json:"status"`
		} `json:"orders"`
	} `json:"data"`
}

type TradeHistoryResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code   int `json:"code"`
		Trades []struct {
			TradeID        int    `json:"trade_id"`
			Pair           string `json:"pair"`
			OrderID        int    `json:"order_id"`
			Side           string `json:"side"`
			Type           string `json:"type"`
			Amount         string `json:"amount"`
			Price          string `json:"price"`
			MakerTaker     string `json:"maker_taker"`
			FeeAmountBase  string `json:"fee_amount_base"`
			FeeAmountQuote string `json:"fee_amount_quote"`
			ExecutedAt     int    `json:"executed_at"`
		} `json:"trades"`
	} `json:"data"`
}

type OrderResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code            int     `json:"code"`
		OrderID         int     `json:"order_id"`
		Pair            string  `json:"pair"`
		Side            string  `json:"side"`
		Type            string  `json:"type"`
		StartAmount     string  `json:"start_amount"`
		RemainingAmount float64 `json:"remaining_amount,string"`
		ExecutedAmount  float64 `json:"executed_amount,string"`
		Price           float64 `json:"price,string"`
		AveragePrice    string  `json:"average_price"`
		OrderedAt       int     `json:"ordered_at"`
		Status          string  `json:"status"`
	} `json:"data"`
}

type CancelOrdersResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code   int `json:"code"`
		Orders []struct {
			OrderID         int     `json:"order_id"`
			Pair            string  `json:"pair"`
			Side            string  `json:"side"`
			Type            string  `json:"type"`
			StartAmount     string  `json:"start_amount"`
			RemainingAmount float64 `json:"remaining_amount,string"`
			ExecutedAmount  float64 `json:"executed_amount,string"`
			Price           float64 `json:"price,string"`
			AveragePrice    string  `json:"average_price"`
			OrderedAt       int     `json:"ordered_at"`
			Status          string  `json:"status"`
		} `json:"orders"`
	} `json:"data"`
}

type CandleResponse struct {
	Success int `json:"success"`
	Data    struct {
		Code        int `json:"code"`
		Candlestick []struct {
			Type  string          `json:"type"`
			Ohlcv [][]interface{} `json:"ohlcv"`
		} `json:"candlestick"`
		Timestamp int64 `json:"timestamp"`
	} `json:"data"`
}

type OrderRequest struct {
	Pair   string  `json:"pair"`
	Amount string  `json:"amount"`
	Price  float64 `json:"price"`
	Side   string  `json:"side"`
	Type   string  `json:"type"`
}

type ActiveOrdersRequest struct {
	Pair  string `url:"pair"`
	Count int    `url:"count"`
}

type TradeHistoryRequest struct {
	Count int    `url:"count"`
	Pair  string `url:"pair"`
}

type CancelOrdersRequest struct {
	Pair     string `json:"pair"`
	OrderIds []int  `json:"order_ids"`
}

type OrdersInfoRequest struct {
	Pair     string `json:"pair"`
	OrderIds []int  `json:"order_ids"`
}

func PostSlack(body string) {
	_, err := resty.R().SetBody(`{"text": "` + body + `"}`).
		Post(SlackWebHookEndpoint)
	if err != nil {
		fmt.Println(err)
	}
}

//使用可能なペアコインを取得します
func GetFreePairCoin() (float64, error) {
	assets, err := GetAssets()
	if err != nil {
		return 0.0, err
	}
	for _, asset := range assets.Data.Assets {
		if asset.Asset == config.CoinPairName {
			return asset.OnhandAmount, nil
		}
	}
	return 0.0, nil
}

//使用可能なコインを取得します。対象コインはconfigのCoinNameを参照します
func GetFreeCoin() (float64, error) {
	assets, err := GetAssets()
	if err != nil {
		return 0.0, err
	}
	for _, asset := range assets.Data.Assets {
		if asset.Asset == config.CoinName {
			return asset.FreeAmount, nil
		}
	}
	return 0.0, nil
}

//保有コインを取得します。対象コインはconfigのCoinNameを参照します
func GetHoldCoin() (float64, error) {
	assets, err := GetAssets()
	if err != nil {
		return 0.0, err
	}
	for _, asset := range assets.Data.Assets {
		if asset.Asset == config.CoinName {
			return asset.OnhandAmount, nil
		}
	}
	return 0.0, nil
}

//Assetsを取得します
func GetAssets() (*AssetsResponse, error) {
	res, err := fetchPrivateAPI(AssetsPath, "GET", nil, &AssetsResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*AssetsResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}

	return res.(*AssetsResponse), nil
}

//未約定の注文を取得します
func GetActiveOrders() (*ActiveOrdersResponse, error) {
	res, err := fetchPrivateAPI(ActiveOrdersPath, "GET", &ActiveOrdersRequest{
		Pair:  config.CoinName + "_" + config.CoinPairName,
		Count: 50000,
	}, &ActiveOrdersResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*ActiveOrdersResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}
	return res.(*ActiveOrdersResponse), nil
}

func GetOrdersInfo(orderIds []int) (*OrdersInfoResponse, error) {

	if len(orderIds) == 0 {
		return &OrdersInfoResponse{}, nil
	}

	res, err := fetchPrivateAPI(OrdersInfoPath, "POST", &OrdersInfoRequest{
		Pair:     config.CoinName + "_" + config.CoinPairName,
		OrderIds: orderIds,
	}, &OrdersInfoResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*OrdersInfoResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}

	return res.(*OrdersInfoResponse), nil
}

//取引履歴を取得します
func GetTradeHistory() (*TradeHistoryResponse, error) {
	res, err := fetchPrivateAPI(TradeHistoryPath, "GET", &TradeHistoryRequest{
		Count: 10,
		Pair:  config.CoinName + "_" + config.CoinPairName,
	}, &TradeHistoryResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*TradeHistoryResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}
	return res.(*TradeHistoryResponse), nil
}

func BuyCoin(amount float64, price float64) (*OrderResponse, error) {
	res, err := fetchPrivateAPI(OrderPath, "POST", &OrderRequest{
		Pair:   config.CoinName + "_" + config.CoinPairName,
		Amount: util.FloatToString(amount),
		Price:  price,
		Side:   "buy",
		Type:   "limit",
	}, &OrderResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*OrderResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}
	fmt.Printf("%fで%f買い注文を入れました\n", price, amount)
	return res.(*OrderResponse), nil
}

func SellCoin(amount float64, price float64) (*OrderResponse, error) {
	res, err := fetchPrivateAPI(OrderPath, "POST", &OrderRequest{
		Pair:   config.CoinName + "_" + config.CoinPairName,
		Amount: util.FloatToString(amount),
		Price:  price,
		Side:   "sell",
		Type:   "limit",
	}, &OrderResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*OrderResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}
	fmt.Printf("%fで%f売り注文を入れました\n", price, amount)

	return res.(*OrderResponse), nil

}

func GetBoard() (*BoardResponse, error) {
	resp, err := resty.R().SetResult(&BoardResponse{}).Get(BoardAPIEndpoint)
	if err != nil {
		return nil, err
	}
	board := new(BoardResponse)
	if err := json.Unmarshal(resp.Body(), board); err != nil {
		return nil, err
	}
	if board.Success != 1 {
		return nil, errors.New(strconv.Itoa(board.Data.Code))
	}
	return board, nil
}

func GetCandle(date time.Time) (*CandleResponse, error) {
	endpoint := CandleAPIEndpointBase + date.Format("20060102")
	resp, err := resty.R().SetResult(&CandleResponse{}).Get(endpoint)
	if err != nil {
		return nil, err
	}
	candle := new(CandleResponse)
	if err := json.Unmarshal(resp.Body(), candle); err != nil {
		return nil, err
	}
	if candle.Success != 1 {
		return nil, errors.New(strconv.Itoa(candle.Data.Code))
	}
	return candle, nil
}

func CancelOrders(orderIds []int) (*CancelOrdersResponse, error) {

	if len(orderIds) == 0 {
		return &CancelOrdersResponse{}, nil
	}

	res, err := fetchPrivateAPI(CancelOrdersPath, "POST", &CancelOrdersRequest{
		Pair:     config.CoinName + "_" + config.CoinPairName,
		OrderIds: orderIds,
	}, &CancelOrdersResponse{})
	if err != nil {
		return nil, err
	}
	resp := res.(*CancelOrdersResponse)
	if resp.Success != 1 {
		return nil, errors.New(strconv.Itoa(resp.Data.Code))
	}

	return res.(*CancelOrdersResponse), nil
}

func fetchPrivateAPI(path string, method string, request interface{}, result interface{}) (interface{}, error) {
	var resp *resty.Response
	var err error
	resty.SetTimeout(time.Duration(30 * time.Second))

	nonce := nonce()
	queryString := ""

	if method == "GET" {
		endPoint := PrivateAPIEndpoint + path
		v, err := query.Values(request)
		if err == nil {
			queryString = v.Encode()
			endPoint += "?" + queryString
		}
		resp, err = resty.R().SetHeader("key", APIKey).
			SetResult(result).
			SetHeader("Content-Type", "application/json").
			SetHeader("ACCESS-KEY", APIKey).
			SetHeader("ACCESS-NONCE", nonce).
			SetHeader("ACCESS-SIGNATURE", signature(method, nonce, path, queryString)).
			Get(endPoint)
	} else {
		body, errJson := json.Marshal(request)
		if errJson != nil {
			fmt.Printf("リクエストが不正です")
			return nil, errJson
		}
		endPoint := PrivateAPIEndpoint + path
		resp, err = resty.R().SetHeader("key", APIKey).
			SetBody(request).
			SetResult(result).
			SetHeader("Content-Type", "application/json").
			SetHeader("ACCESS-KEY", APIKey).
			SetHeader("ACCESS-NONCE", nonce).
			SetHeader("ACCESS-SIGNATURE", signature(method, nonce, path, string(body))).
			Post(endPoint)
	}

	if err != nil {
		return nil, err
	}
	return resp.Result(), nil
}

//プライベートAPI呼び出しに必要な��通の���クエストパラ���ー���������������取得します
func nonce() string {
	nonceIncrementalCounter += 1.0
	nonce := strconv.FormatFloat(float64(time.Now().Unix()*10000)+nonceIncrementalCounter, 'f', 0, 64)
	return nonce
}

//queryStringの署名�������字列を返却します
func signature(method string, nonce string, path string, queryStringOrBody string) string {
	if apiSecret.Value == "" {
		raw, err := ioutil.ReadFile(config.SecJsonFileName)
		if err != nil {
			panic("Secの読み込みに失敗しました")

		}
		json.Unmarshal(raw, &apiSecret)
	}
	rawString := ""
	if method == "GET" {
		rawString = nonce + path
		if queryStringOrBody != "" {
			rawString += "?" + queryStringOrBody
		}
	} else {
		rawString = nonce + queryStringOrBody
	}
	mac := hmac.New(sha256.New, []byte(apiSecret.Value))
	mac.Write([]byte(rawString))
	return hex.EncodeToString(mac.Sum(nil))
}
