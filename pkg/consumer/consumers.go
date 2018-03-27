// websocket response consumers from the gdax websocket api https://docs.gdax.com/#subscribe
// https://godoc.org/github.com/preichenberger/go-gdax#Message
package consumer

import (
	"github.com/preichenberger/go-gdax"
	"github.com/jinzhu/gorm"
	db "CIP-exchange-consumer-gdax/internal/db"
	"strconv"
	"time"
	"log"
)

//main consumer, spawning a consumer for each trading pair
func MessageConsumer(message gdax.Message, gorm gorm.DB, WorkerPerPair int, chanGuide map[string](chan gdax.Message)){

	// start a new workerset for a singe pair
	if message.Type == "snapshot"{
		market := db.CreateMarket(gorm, message.ProductId[0:3], message.ProductId[len(message.ProductId)-3:])
		orderbook := db.CreateOrderbook(gorm, market)

		// try to close the old channel to stop old workers
		if _, ok := chanGuide[message.ProductId]; ok {
			close(chanGuide[message.ProductId])
		}

		// create a new channel for the pair
		workerChan := make(chan gdax.Message, 100)
		chanGuide[message.ProductId] = workerChan

		// start a number of worker processes
		for i := 1; i <= WorkerPerPair; i++{
			worker := Worker{gorm, orderbook, market, workerChan}
			go worker.start()
		}

	}

//	put the message in the appropriate worker channel for consumption
	chanGuide[message.ProductId] <- message
}

//
type Worker struct{
	gorm gorm.DB
	orderbook *db.GdaxOrderBook
	market *db.GdaxMarket
	messages <- chan gdax.Message
}
func (w Worker) start (){
	for true {
		message := <-w.messages


		if message.Type == "ticker"{
			//message.Time.Time() is unreliable, so we have to use our own time
			db.AddTicker(w.gorm, w.market, message.BestBid, message.BestAsk, time.Now())
		}

		// parsing initial data (only performed on worker startup
		if message.Type == "snapshot" {
			for _, ask := range message.Asks{
				price, err := strconv.ParseFloat(ask[0], 64)
				if err != nil{
					log.Panic(err)
				}
				quantity, err := strconv.ParseFloat(ask[1], 64)
				if err != nil{
					log.Panic(err)
				}
				db.AddOrder(w.gorm, *w.orderbook, false, price, quantity)
			}
			for _, ask := range message.Bids{
				price, err := strconv.ParseFloat(ask[0], 64)
				if err != nil{
					log.Panic(err)
				}
				quantity, err := strconv.ParseFloat(ask[1], 64)
				if err != nil{
					log.Panic(err)
				}
				db.AddOrder(w.gorm, *w.orderbook, true, price, quantity)
			}
		}

		if message.Type == "l2update" {
			for _, change := range message.Changes{
				var buy = false

				if change[0] == "buy"{
					buy = true
				}

				price, err := strconv.ParseFloat(change[1], 64)
				if err != nil{
					log.Panic(err)
				}
				quantity, err := strconv.ParseFloat(change[2], 64)
				if err != nil{
					log.Panic(err)
				}
				db.AddOrder(w.gorm, *w.orderbook, buy, price, quantity)
			}
		}
	}
}
