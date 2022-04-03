package main

import (
	"fmt"
	"time"
)

func updateOrderbookFromLiveOrder(orderbooks Orderbooks, liveorders LiveOrders) {
	if liveorders.BTCUSD.Microtimestamp > orderbooks.BTCUSD.Microtimestamp {
		fmt.Printf(" +++ Apply LiveOrders: %+v\n", liveorders)
	}

}

func latestWebsocketOrderbook(OrderbookChannel chan Orderbooks) (orderbooks Orderbooks) {
	for {
		fmt.Printf("Processing orderbook channel queue: %+v\n", len(OrderbookChannel))
		orderbooks = <-OrderbookChannel
		if len(OrderbookChannel) == 0 {
			fmt.Printf("Orderbooks: %+v\n", orderbooks)
			return orderbooks
		}
	}

}

func main() {
	LiveOrdersChannel := make(chan LiveOrders, 1000)
	OrderbookChannel := make(chan Orderbooks, 10)
	UnsubscribeOrderbookChannel := make(chan UnsubscribeOrderbook)

	go WebsocketOrderbook(OrderbookChannel, UnsubscribeOrderbookChannel)
	go WebsocketLiveOrders(LiveOrdersChannel)

	for {
		time.Sleep(100 * time.Millisecond)
		if len(OrderbookChannel) > 1 {
			break
		}
	}

	UnsubscribeOrderbookChannel <- UnsubscribeOrderbook{
		Unsubscribe: "Unsubscribe",
	}

	orderbooks := latestWebsocketOrderbook(OrderbookChannel)

	go func() {
		for {
			select {
			case liveorders := <-LiveOrdersChannel:
				updateOrderbookFromLiveOrder(orderbooks, liveorders)
			}
		}
	}()

	time.Sleep(1 * time.Second)
}
