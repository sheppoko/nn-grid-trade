package config

const (
	Debug = 0
)

const (
	CoinName                     = "btc"
	CoinPairName                 = "jpy"
	OrderNumInOnetime            = 5
	UnSoldBuyPositionLogFileName = "unsold_buy_position.json"
	SecJsonFileName              = "sec.json"
)

var (
	BuyRange               = 0.018
	TakeProfitRange        = 0.018
	MaxPositionCount       = 29
	PositionMaxDownPercent = 40.0
)
