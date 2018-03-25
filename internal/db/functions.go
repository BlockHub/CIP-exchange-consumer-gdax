package db

import (
	"github.com/jinzhu/gorm"
	"time"
	"strings"
)

func CreateOrderbook (gorm gorm.DB, market *GdaxMarket) *GdaxOrderBook{
	orderbook := GdaxOrderBook{0, market.ID, time.Now()}
	err := gorm.Create(&orderbook).Error
	if err != nil{
		panic(err)
	}
	return &orderbook
}

func CreateMarket(gorm gorm.DB, ticker string, quote string) *GdaxMarket {
	market := GdaxMarket{0, ticker, quote}
	err := gorm.Create(&market).Error

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			gorm.Where(map[string]interface{}{"ticker": ticker, "quote": quote}).Find(&market)
		}
	}
	return &market
}

func AddOrder(gorm gorm.DB, book GdaxOrderBook, buy bool, rate float64, quantity float64, ){
	order := GdaxOrder{0, book.ID, buy, rate, quantity, time.Now()}
	err := gorm.Create(&order).Error
	if err != nil{
		panic(err)
	}
}

func AddTicker(gorm gorm.DB, market *GdaxMarket, bestBid float64, bestAsk float64, time time.Time){
	ticker := GdaxTicker{0, market.ID, bestBid, bestAsk, time}
	err := gorm.Create(&ticker).Error
	if err != nil{
		panic(err)
	}
}