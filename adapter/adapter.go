package adapter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"nn-grid-trade/api"
	"nn-grid-trade/config"
	"nn-grid-trade/util"
	"time"
)

//
type UnsoldBuyOrder struct {
	OrderID            int     `json:"order_id"`
	BuyPrice           float64 `json:"buy_price"`
	RemainingBuyAmount float64 `json:"remaining_buy_amount"`
}

//キャンセルしてしまった売りオーダ
type CanceldSellOrder struct {
	Price  float64 `json:"buy_price"`
	Amount float64 `json:"amount"`
}

var unSoldBuyOrders = []*UnsoldBuyOrder{}
var latestPositionNum int
var activeOrdersCache *api.ActiveOrdersResponse
var highestMarketPrice = 0.0

func StartStrategy() {
	postionNum := 0
	counter := 0
	for {
		time.Sleep(1000 * time.Millisecond) // 休む
		initCache()
		if counter%60 == 0 && false {
			counter = 0
			errCandle := SetRangeFromCandle()
			if errCandle != nil {
				fmt.Println("ロウソク取得中にエラーが発生しました")
				fmt.Println(errCandle)
				continue
			}
		}
		counter++
		if counter%600 == 1 {
			counter = 1
			err := PostInfoToSlack()
			if err != nil {
				fmt.Println(err)
			}
		}

		_, err := SellCoinIfNeedAndUpdateUnsold()
		if err != nil {
			fmt.Println("売り注文必要チェック及び売り注文作成中にエラーが発生しました")
			fmt.Println(err)
			continue
		}
		//ポジション数を取得
		positionNumTmp, err := GetSellPriceKindNum()
		if err != nil {
			fmt.Println("ポジション数取得時にエラーが発生しました")
			fmt.Println(err)
			continue
		}
		//ポジション数変化し、ポジション0の場合は全ての買い注文キャンセルして発注しなおす
		if positionNumTmp != postionNum && positionNumTmp == 0 {
			CancelAllBuyOrders()
		}
		postionNum = positionNumTmp
		//一番安い売り注文から2段階下げた買い注文を起点に5個入れる
		//1つずつ値計算し、それより高いか同値の注文がある場合は飛ばす
		err = OrderIfNeed(positionNumTmp)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func initCache() {
	activeOrdersCache = nil
}

func SetRangeFromCandle() error {
	for lo := 20.0; lo < 90; lo = lo + 5 {
		for st := -6; st > -7; st = st - 1 {
			config.PositionMaxDownPercent = lo
			dateNum := 5    //シミュレーション日数
			startDiff := st //開始日（本日起点）
			baseDateDiff := startDiff
			candle, _ := api.GetCandle(time.Now().AddDate(0, 0, baseDateDiff))
			for i := 1; i < dateNum; i++ {
				candlePart, e := api.GetCandle(time.Now().AddDate(0, 0, baseDateDiff+i))
				//fmt.Println(time.Now().AddDate(0, 0, baseDateDiff+i))
				if e != nil {
					fmt.Println(e)
				}
				//fmt.Println("recieve candle", i+1)
				candle.Data.Candlestick[0].Ohlcv = append(candle.Data.Candlestick[0].Ohlcv, candlePart.Data.Candlestick[0].Ohlcv...)
			}

			bestRange := 0.0
			bestMaxPosition := 0.0
			//bestTpCounters := []int{}
			bestTakeProfitCounter := 0.0
			bestCounter := 0
			songiriCounter := 0
			for tp := 0.001; tp < 0.05; tp = tp + 0.001 {

				buyRange := tp
				songiriCounter = 0
				maxHigh := 0.0
				basePrice, _ := util.StringToFloat(candle.Data.Candlestick[0].Ohlcv[0][2].(string))
				positions := []float64{basePrice}
				maxPosition := MaxPositionFromRange(buyRange) - 1
				money := 1000000.0
				tpCounters := []int{}
				tpCounter := 0
				counter := 0
				profitPercentPerTime := (1 / maxPosition) * tp

				for j, data := range candle.Data.Candlestick[0].Ohlcv {
					round := j % (24 * 60)
					if round == 24*60-1 {
						tpCounters = append(tpCounters, tpCounter)
						tpCounter = 0
					}
					high, _ := util.StringToFloat(data[1].(string))
					low, _ := util.StringToFloat(data[2].(string))

					if maxHigh < high {
						maxHigh = high
					}

					if maxHigh*(100-config.PositionMaxDownPercent)/100 > low {
						// av := average(positions)
						// money *= low * 0.9984 / av
						positions = []float64{low}
						// fmt.Println("損切り", low, "追加")
						maxHigh = low
						songiriCounter++
					}
					hajime, _ := util.StringToFloat(data[0].(string))
					owari, _ := util.StringToFloat(data[3].(string))
					isYosen := hajime < owari
					nPos := positions

					if isYosen {
						start := lowest(positions) * (1 - buyRange)
						add := false
						for i := start; i > low; i = i * (1 - buyRange) {
							//fmt.Println(low, i, "追加")
							nPos = append(nPos, i)
							add = true

						}
						positions = nPos
						if add {
							//fmt.Println(low)
						}
						sell := false
						for _, p := range positions {
							if p*(1+tp) < high {
								// fmt.Println(high, p, "売却")
								tpCounter++
								money += 1000000 * (profitPercentPerTime + 0.0000/maxPosition)
								counter++
								sell = true
								nPos = remove(nPos, p)
							}
						}

						if sell {
							// fmt.Println(high)
						}
						positions = nPos
						if len(positions) == 0 {
							// fmt.Println("成行", high, "追加")
							positions = append(positions, high*1.01)
						}
					} else {
						nPos = positions
						for _, p := range positions {
							if p*(1+tp) < high {
								// fmt.Println(high, p, "売却")
								tpCounter++
								money += 1000000 * (profitPercentPerTime + 0.0000/maxPosition)
								counter++
								nPos = remove(positions, p)
							}
						}
						positions = nPos
						if len(positions) == 0 {
							// fmt.Println("成行", high, "追加")
							positions = append(positions, high*1.01)
						}
						start := lowest(positions) * (1 - buyRange)
						for i := start; i > low; i = i * (1 - buyRange) {
							// fmt.Println(low, i, "追加")
							positions = append(positions, i)
						}
					}
					// if ((high-basePrice)/basePrice)*0.75 > buyRange && isYosen {
					// 	simCounter += float64(int64(((high - basePrice) / basePrice) * 0.75 / buyRange))
					// 	money = money * (1 + profitPercentPerTime)
					// 	shouldUpdateLow = true
					// }
				}

				if money > bestTakeProfitCounter {
					//bestTpCounters = tpCounters
					bestRange = tp
					bestMaxPosition = maxPosition
					bestCounter = counter
					bestTakeProfitCounter = money
				}
			}
			profitPerTime := 1.0 + (float64(bestTakeProfitCounter)-1000000.0)/float64(bestCounter)/1000000
			fukuri := math.Pow(profitPerTime, float64(bestCounter))
			// fukuriMoney := 1000000 * fukuri
			yearFukuri := math.Pow(fukuri, 365/float64(dateNum))
			//util.PrettyPrint(bestTpCounters)
			fmt.Print(time.Now().AddDate(0, 0, baseDateDiff).Format("2006-01-02"))
			fmt.Printf("から%d日間でのシミュレーション（", dateNum)
			fmt.Printf("%v％耐え設定）\n", lo)
			fmt.Printf("ポジション数:%f,最適レンジ：%f,1年耐えきった時の年間期待倍率:%f,実績損切り回数:%d\n", bestMaxPosition, bestRange, yearFukuri, songiriCounter)
		}
	}
	return nil

}

func remove(numbers []float64, search float64) []float64 {
	success := false
	result := []float64{}
	for _, num := range numbers {
		if num != search {
			result = append(result, num)
		} else {
			success = true
		}
	}
	if !success {
		fmt.Println("削除失敗")
	}
	return result
}

func average(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum / float64(len(numbers))
}

func lowest(numbers []float64) float64 {
	tempLow := 1234567890.0
	for _, num := range numbers {
		if num < tempLow {
			tempLow = num
		}
	}
	if tempLow == 1234567890.0 {
		tempLow = 0
	}
	return tempLow
}

func MaxPositionFromRange(buyRange float64) float64 {
	pricePer := 100.0
	positionNum := 0.0
	for {
		positionNum++
		pricePer *= 1 - buyRange
		if (100.0 - pricePer) > config.PositionMaxDownPercent {
			break
		}
	}
	return positionNum
}

func LoadUnSoldStatus() (bool, error) {
	raw, err := ioutil.ReadFile(config.UnSoldBuyPositionLogFileName)
	if err != nil {
		fmt.Println("状態の復元に失敗しました")
		return false, err
	} else {
		err = json.Unmarshal(raw, &unSoldBuyOrders)
		fmt.Println("状態を復元しました")
	}
	return true, nil
}

func PostInfoToSlack() error {
	pair, err := api.GetFreePairCoin()
	if err != nil {
		return err
	}
	pairEstimate, errEstimate := GetMoneyIfAllSellEstablish()
	if errEstimate != nil {
		return errEstimate
	}

	coin, errCoin := api.GetHoldCoin()
	if errCoin != nil {
		return errCoin
	}
	board, errBoard := api.GetBoard()
	if errBoard != nil {
		return errBoard
	}
	boardPrice, errFloat := util.StringToFloat(board.Data.Bids[0][0])
	if errFloat != nil {
		return errFloat
	}
	estimate := coin*boardPrice + pair
	api.PostSlack("現在資産:" + util.FloatToString(estimate) + ",全利益確定時:" + util.FloatToString(pairEstimate+pair))
	return nil
}

//一番安い売り注文か現在最良Askの高い方から2段階下げた買い注文を起点に5個入れます
//1つずつ値計算し、類似価格の注文がある場合は飛ばします
func OrderIfNeed(nowPositionNum int) error {
	lowestSell, err := GetLowestSellOrderPrice()
	if err != nil {
		return err
	}
	board, errB := api.GetBoard()
	if errB != nil {
		return errB
	}
	boardPrice, _ := util.StringToFloat(board.Data.Asks[0][0])
	startPrice := (lowestSell / (1 + config.TakeProfitRange)) * (1 - config.BuyRange)
	if lowestSell < boardPrice {
		startPrice = (boardPrice / (1 + config.TakeProfitRange))
	}
	if nowPositionNum == 0 {
		startPrice = boardPrice
	}
	orderNum := config.MaxPositionCount - nowPositionNum
	if orderNum == 0 {
		return nil
	}
	buyMax := orderNum
	if buyMax >= config.OrderNumInOnetime {
		buyMax = config.OrderNumInOnetime
	}
	for i := 0; i < buyMax; i++ {
		p := startPrice * math.Pow((1-config.BuyRange), float64(i))
		hasRangeBuyOrder, err := hasRangeBuyOrder(p)
		if err != nil {
			return err
		}
		if !hasRangeBuyOrder {
			freePair, errPair := api.GetFreePairCoin()
			if errPair != nil {
				return errPair
			}
			usePair := freePair / float64(orderNum)
			amount := usePair / p
			_, errBuy := BuyCoinAndRegistUnsold(amount, p)
			if errBuy != nil {
				return errBuy
			}
		} else {
			continue
		}
	}
	return nil
}

func hasRangeBuyOrder(price float64) (bool, error) {
	activeOrder, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return false, err
	}
	for _, order := range activeOrder.Data.Orders {
		if order.Side == "buy" && (math.Abs(order.Price-price) < price*config.BuyRange) {
			return true, nil
		}
	}
	return false, nil
}

func BuyCoinAndRegistUnsold(amount float64, price float64) (bool, error) {
	amount = util.Round(amount, 4)

	res, err := api.BuyCoin(amount, price)
	if err != nil {
		return false, err
	}
	appendOrder := new(UnsoldBuyOrder)
	appendOrder.OrderID = res.Data.OrderID
	appendOrder.BuyPrice = res.Data.Price
	appendOrder.RemainingBuyAmount = res.Data.RemainingAmount
	unSoldBuyOrders = append(unSoldBuyOrders, appendOrder)
	util.SaveJsonToFile(unSoldBuyOrders, config.UnSoldBuyPositionLogFileName)

	return true, nil
}

//売り注文が出されていない買い注文と現在の状況をチェック�����必要であれば売り注文をなげます。
//戻り値��約定した事が判明し������い注文のうちもっとも低い価��の注文です
func SellCoinIfNeedAndUpdateUnsold() (float64, error) {
	max := 1234567890.0
	lowestPrice := max
	orderIds := []int{}
	var res = &api.OrdersInfoResponse{}

	for i, order := range unSoldBuyOrders {
		orderIds = append(orderIds, order.OrderID)
		if i%10 == 0 || i == len(unSoldBuyOrders)-1 {
			resPart, err := api.GetOrdersInfo(orderIds)
			if err != nil {
				return -1.0, err
			}
			res.Data.Orders = append(res.Data.Orders, resPart.Data.Orders...)
			orderIds = []int{}
		}
	}

	for _, order := range res.Data.Orders {
		for _, unSold := range unSoldBuyOrders {
			//当該注��の残量が��化していた場合は対応��た売り注文を投げる
			if unSold.OrderID == order.OrderID && unSold.RemainingBuyAmount != order.RemainingAmount {

				//一部約���の時に注���������通らない
				sellAmount := unSold.RemainingBuyAmount - order.RemainingAmount
				sellPrice := unSold.BuyPrice * (1 + config.TakeProfitRange)
				fmt.Println("買���注文が約定しているため売り注文を作成し���す...")
				if order.Price < lowestPrice {
					lowestPrice = order.Price
				}
				_, err := api.SellCoin(sellAmount, sellPrice)
				if err != nil {
					fmt.Println(unSold.RemainingBuyAmount, order.RemainingAmount)
					return -1, err
				}

				if order.RemainingAmount <= 0 {
					DeleteUnSoldOrder(order.OrderID)
				} else {
					unSold.RemainingBuyAmount -= sellAmount
				}

				util.SaveJsonToFile(unSoldBuyOrders, config.UnSoldBuyPositionLogFileName)
				fmt.Println("作成しま����た")
			}
		}
	}
	if lowestPrice == max {
		lowestPrice = -1
	}

	return lowestPrice, nil
}

func DeleteUnSoldOrder(orderId int) []*UnsoldBuyOrder {
	orders := []*UnsoldBuyOrder{}
	for _, unSoldOrder := range unSoldBuyOrders {
		if unSoldOrder.OrderID != orderId {
			orders = append(orders, unSoldOrder)
		}
	}
	unSoldBuyOrders = orders
	return orders
}

func DeleteAllUnSoldOrder() {
	orders := []*UnsoldBuyOrder{}
	unSoldBuyOrders = orders
}

//売り買い全ての注文をキャンセルします
func CancelAllOrders() (bool, error) {
	targetOrderId := []int{}
	orders, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return false, err
	}
	for _, order := range orders.Data.Orders {
		targetOrderId = append(targetOrderId, order.OrderID)
	}
	_, err = api.CancelOrders(targetOrderId)
	if err != nil {
		return false, err
	}
	DeleteAllUnSoldOrder()
	util.SaveJsonToFile(unSoldBuyOrders, config.UnSoldBuyPositionLogFileName)
	fmt.Println("全ての注文をキャンセルしました")
	return true, nil
}

func GetLowestSellOrderPrice() (float64, error) {

	max := 123456789.00
	lowestSell := max
	activeOrder, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return -1.0, err
	}
	for _, order := range activeOrder.Data.Orders {
		if order.Side == "sell" {
			if order.Price < lowestSell {
				lowestSell = order.Price
			}
		}
	}
	if lowestSell == max {
		return 0, nil
	}
	return lowestSell, nil
}

//売り注文の��段���種類数��������������取����������す
func GetSellPriceKindNum() (int, error) {
	res, err := GetActiveOrdersFromAPIorCache()
	prices := []float64{}
	if err != nil {
		return 0, err
	}
	for _, order := range res.Data.Orders {
		if order.Side == "sell" {
			shouldAdd := true
			for _, price := range prices {
				if price == order.Price {
					shouldAdd = false
				}
			}
			if shouldAdd {
				prices = append(prices, order.Price)
			}
		}
	}
	return len(prices), nil
}

func GetBuyOrderNum() (int, error) {
	ret := 0
	res, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return 0, err
	}
	for _, order := range res.Data.Orders {
		if order.Side == "buy" {
			ret++
		}
	}
	return ret, nil
}
func GetSellOrderNum() (int, error) {
	ret := 0
	res, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return 0, err
	}
	for _, order := range res.Data.Orders {
		if order.Side == "sell" {
			ret++
		}
	}
	return ret, nil
}

func GetMoneyIfAllSellEstablish() (float64, error) {
	ret := 0.0
	res, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return 0, err
	}
	for _, order := range res.Data.Orders {
		if order.Side == "sell" {
			ret += order.RemainingAmount * order.Price
		}
	}
	return ret, nil
}

//買い注文を全てキャンセルします。TODO:100件上限しかキャンセルできてない
func CancelAllBuyOrders() (bool, error) {
	targetOrderId := []int{}
	orders, err := GetActiveOrdersFromAPIorCache()
	if err != nil {
		return false, err
	}
	counter := 0
	for _, order := range orders.Data.Orders {
		if order.Side == "buy" {
			targetOrderId = append(targetOrderId, order.OrderID)
		}
		counter++
		if counter == 100 {
			break
		}
	}
	_, errCancel := api.CancelOrders(targetOrderId)
	if errCancel != nil {
		return false, errCancel
	}
	DeleteAllUnSoldOrder()
	util.SaveJsonToFile(unSoldBuyOrders, config.UnSoldBuyPositionLogFileName)
	fmt.Println("全ての買い注文をキ�����ンセ�������������した")
	return true, nil
}

func GetActiveOrdersFromAPIorCache() (*api.ActiveOrdersResponse, error) {
	if activeOrdersCache == nil {
		res, err := api.GetActiveOrders()
		if err != nil {
			return nil, err
		}
		activeOrdersCache = res
		return res, nil
	}

	return activeOrdersCache, nil
}
