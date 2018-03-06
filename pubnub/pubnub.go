package pubnub

import (
	"bitbank-grid-trade/adapter"
	"bitbank-grid-trade/api"
	"bitbank-grid-trade/config"
	"bitbank-grid-trade/util"
	"encoding/json"
	"fmt"
	"math"

	"github.com/pubnub/go/messaging"
)

func StartSubscribeBoard() {
	pubnub := messaging.NewPubnub(
		"", "sub-c-e12e9174-dd60-11e6-806b-02ee2ddab7fe",
		"", "", false, "", nil)

	channel := "depth_" + config.CoinName + "_jpy"
	sucCha := make(chan []byte)
	errCha := make(chan []byte)

	go pubnub.Subscribe(channel, "", sucCha, false, errCha)

	go func() {
		for {
			select {
			case res := <-sucCha:
				var msg []interface{}
				err := json.Unmarshal(res, &msg)
				if err != nil {
					fmt.Println("板情報受信中に不正なjsonを受信しました")
					continue
				}
				switch msg[0].(type) {
				case float64:
				case []interface{}:

					board := new(api.BoardResponse)
					message := msg[0].([]interface{})[0].(map[string]interface{})
					obj, errMarshal := json.Marshal(message)
					if errMarshal != nil {
						fmt.Println("板情報受信中に不正なjsonを受信しました")
						continue
					}
					err = json.Unmarshal(obj, &board)
					if err != nil {
						fmt.Println("板情報受信中に不正なjsonを受信しました")
					}
					adapter.LatestBoard = board
					adapter.BoardUpdated()

				default:
					fmt.Println("板情報受信中に不正なjsonを受信しました")
				}
			case err := <-errCha:
				fmt.Println(string(err))
			case <-messaging.SubscribeTimeout():
				fmt.Println("板情報受信がタイムアウトしました")
			}
		}
	}()
}

func StartSubscribeMarket() {
	pubnub := messaging.NewPubnub(
		"", "sub-c-e12e9174-dd60-11e6-806b-02ee2ddab7fe",
		"", "", false, "", nil)

	channel := "transactions_" + config.CoinName + "_jpy"
	sucCha := make(chan []byte)
	errCha := make(chan []byte)

	go pubnub.Subscribe(channel, "", sucCha, false, errCha)

	go func() {
		lastSide := "buy"
		lastPrice := 0.0
		priceDiff := 0.0
		counter := 0.0
		for {
			select {
			case res := <-sucCha:
				var msg []interface{}
				err := json.Unmarshal(res, &msg)
				if err != nil {
					fmt.Println("板情報受信中に不正なjsonを受信しました")
					continue
				}
				switch msg[0].(type) {
				case float64:
				case []interface{}:
					message := msg[0].([]interface{})[0].(map[string]interface{})
					data := message["data"].(map[string]interface{})
					transaction := data["transactions"].([]interface{})[0]
					price := transaction.(map[string]interface{})["price"]
					side := transaction.(map[string]interface{})["side"]
					pricef, _ := util.StringToFloat(price.(string))
					if lastSide != side.(string) {
						if lastPrice != 0.0 {
							priceDiff += math.Abs(pricef - lastPrice)
							lastSide = side.(string)
							counter++
							fmt.Println(priceDiff/counter, math.Abs(pricef-lastPrice))

						}
					}
					lastPrice = pricef

				default:
					fmt.Println("板情報受信中に不正なjsonを受信しました")
				}
			case err := <-errCha:
				fmt.Println(string(err))
			case <-messaging.SubscribeTimeout():
				fmt.Println("板情報受信がタイムアウトしました")
			}
		}
	}()
}
