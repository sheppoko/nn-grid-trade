package main

import (
	"fmt"
	"nn-grid-trade/adapter"
)

func main() {

	_, err := adapter.LoadUnSoldStatus()
	if err != nil {
		fmt.Println(err)
		return
	}
	adapter.StartStrategy()

}
