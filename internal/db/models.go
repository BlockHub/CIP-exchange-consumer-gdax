package db

import "time"

type GdaxOrderBook struct {
	ID uint64 			`gorm:"primary_key"`
	MarketID uint64
	Time time.Time		`gorm:"primary_key"`
}

type GdaxOrder struct {
	ID uint64 			`gorm:"primary_key"`
	OrderbookID uint64
	Buy bool
	Rate float64
	//bitfinex supports giving the total number of sell/buyorders.
	//however we should skimp on memory and not add those
	//count int64
	Quantity float64
	Time time.Time		`gorm:"primary_key"`
}

type GdaxMarket struct {
	ID uint64 			`gorm:"primary_key"`
	Ticker string		`gorm:"unique_index:idx_market"`
	Quote string		`gorm:"unique_index:idx_market"`
}

type GdaxTicker struct {
	ID  uint64 			`gorm:"primary_key"`
	MarketID uint64
	BestBid float64
	BestAsk float64
	Time time.Time		`gorm:"primary_key"`
}

type GdaxTrade struct {
	ID  uint64 			`gorm:"primary_key"`
	MarketID uint64
	Size float64
	Price float64
	Buy bool
	Time time.Time		`gorm:"primary_key"`
}
