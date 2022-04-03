package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bitstonks/bitstamp-go"
	"github.com/shopspring/decimal"
)

// Control channel
type UnsubscribeOrderbook struct {
	Unsubscribe string
}

// Orderbooks contains multiple orderbook strucsts
type Orderbooks struct {
	BTCUSD Orderbook
}

// Orderbook contains top asks, bids and timestamp
// from latest websocket event
type Orderbook struct {
	Microtimestamp int
	Ask            PriceAmount
	Bid            PriceAmount
}

// PriceAmount contains Price and Amount fields
// from latest websocket event
type PriceAmount struct {
	Price  decimal.Decimal
	Amount decimal.Decimal
}

func updateOrderbook(event *bitstamp.WsEvent) (channel string, orderbook Orderbook, ok bool) {
	channel = (*event).Channel

	orderbook, ok = parseOrderbook((*event).Data)
	if !ok {
		return "", Orderbook{}, false
	}

	return channel, orderbook, true
}

func parseOrderbook(data interface{}) (Orderbook, bool) {
	DataMap, ok := data.(map[string]interface{})
	if !ok {
		return Orderbook{}, false
	}

	MicrotimestampInt, ok := parseOrderbookTimestamp(DataMap)
	if !ok {
		return Orderbook{}, false
	}

	AsksPriceAmount, ok := parseOrderbookAsks(DataMap)
	if !ok {
		return Orderbook{}, false
	}

	BidsPriceAmount, ok := parseOrderbookBids(DataMap)
	if !ok {
		return Orderbook{}, false
	}

	return Orderbook{
		Ask:            AsksPriceAmount,
		Bid:            BidsPriceAmount,
		Microtimestamp: MicrotimestampInt,
	}, true
}

func parseOrderbookTimestamp(DataMap map[string]interface{}) (int, bool) {
	MicrotimestampString, ok := DataMap["microtimestamp"].(string)
	if !ok {
		return 0, false
	}

	MicrotimestampInt, err := strconv.Atoi(MicrotimestampString)
	if err != nil {
		return 0, false
	}

	return MicrotimestampInt, true
}

func parseOrderbookAsks(DataMap map[string]interface{}) (PriceAmount, bool) {
	Asks, ok := DataMap["asks"].([]interface{})
	if !ok {
		return PriceAmount{}, false
	}

	AsksPriceAmount, ok := parseOrderbookTopPriceAmount(Asks[0])
	if !ok {
		return PriceAmount{}, false
	}

	return AsksPriceAmount, true
}

func parseOrderbookBids(DataMap map[string]interface{}) (PriceAmount, bool) {
	Bids, ok := DataMap["bids"].([]interface{})
	if !ok {
		return PriceAmount{}, false
	}

	AsksPriceAmount, ok := parseOrderbookTopPriceAmount(Bids[0])
	if !ok {
		return PriceAmount{}, false
	}

	return AsksPriceAmount, true
}

func parseOrderbookTopPriceAmount(top interface{}) (PriceAmount, bool) {
	priceAmount, ok := top.([]interface{})
	if !ok {
		return PriceAmount{}, false
	}

	priceString, ok := priceAmount[0].(string)
	if !ok {
		return PriceAmount{}, false
	}

	priceDecimal, err := decimal.NewFromString(priceString)
	if err != nil {
		return PriceAmount{}, false
	}

	amountString, ok := priceAmount[1].(string)
	if !ok {
		return PriceAmount{}, false
	}

	amountDecimal, err := decimal.NewFromString(amountString)
	if err != nil {
		return PriceAmount{}, false
	}

	return PriceAmount{
		Price:  priceDecimal,
		Amount: amountDecimal,
	}, true
}

// WebsocketOrderbook returns latest top ask/bids from websocket events
func WebsocketOrderbook(orderbookChannel chan Orderbooks, UnsubscribeOrderbookChannel chan UnsubscribeOrderbook) {

	ws, err := bitstamp.NewWsClient()
	if err != nil {
		log.Panicf("error initializing client %v", err)
	}

	ws.Subscribe(
		"order_book_btcusd",
	)

	ob := map[string]Orderbook{}

	for {
		select {
		case ev := <-ws.Stream:
			channel, orderbook, ok := updateOrderbook(ev)

			if ok {
				ob[channel] = orderbook
			}

			orderbookChannel <- Orderbooks{
				BTCUSD: ob["order_book_btcusd"],
			}

		case err := <-ws.Errors:
			fmt.Printf("--- ERROR: %#v\n", err)

		case <-UnsubscribeOrderbookChannel:
			ws.Unsubscribe(
				"order_book_btcusd",
			)
		}
	}

}
