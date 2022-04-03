package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bitstonks/bitstamp-go"
	"github.com/shopspring/decimal"
)

type LiveOrders struct {
	BTCUSD LiveOrder
}

type LiveOrder struct {
	Microtimestamp int
	Type           int
	Price          decimal.Decimal
	Amount         decimal.Decimal
}

func processOrder(event *bitstamp.WsEvent) (channel string, order LiveOrder, ok bool) {
	channel = (*event).Channel

	order, ok = parseLiveOrder((*event).Data)
	if !ok {
		return "", LiveOrder{}, false
	}

	return channel, order, true
}

func parseLiveOrder(data interface{}) (LiveOrder, bool) {
	DataMap, ok := data.(map[string]interface{})
	if !ok {
		return LiveOrder{}, false
	}

	microtimestampString, ok := DataMap["microtimestamp"].(string)
	if !ok {
		return LiveOrder{}, false
	}

	MicrotimestampInt, err := strconv.Atoi(microtimestampString)
	if err != nil {
		return LiveOrder{}, false
	}

	priceString, ok := DataMap["price"].(float64)
	if !ok {
		return LiveOrder{}, false
	}

	priceDecimal := decimal.NewFromFloat(priceString)
	if err != nil {
		return LiveOrder{}, false
	}

	amountString, ok := DataMap["amount"].(float64)
	if !ok {
		return LiveOrder{}, false
	}

	amountDecimal := decimal.NewFromFloat(amountString)
	if err != nil {
		return LiveOrder{}, false
	}

	typeFloat, ok := DataMap["order_type"].(float64)
	if !ok {
		return LiveOrder{}, false
	}

	typeInt := int(typeFloat)
	if err != nil {
		return LiveOrder{}, false
	}

	return LiveOrder{
		Price:          priceDecimal,
		Amount:         amountDecimal,
		Type:           typeInt,
		Microtimestamp: MicrotimestampInt,
	}, true
}

// WebsocketLiveOrders
func WebsocketLiveOrders(liveordersChannel chan LiveOrders) {

	ws, err := bitstamp.NewWsClient()
	if err != nil {
		log.Panicf("error initializing client %v", err)
	}

	ws.Subscribe(
		"live_orders_btcusd",
	)

	order := map[string]LiveOrder{}

	for {
		select {
		case ev := <-ws.Stream:
			channel, liveorder, ok := processOrder(ev)

			if ok {
				order[channel] = liveorder
			}

			liveordersChannel <- LiveOrders{
				BTCUSD: order["live_orders_btcusd"],
			}

		case err := <-ws.Errors:
			fmt.Printf("--- ERROR: %#v\n", err)

		}
	}

}
