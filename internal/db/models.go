package db

type GdaxOrderBook struct {
	ID uint 			`gorm:"primary_key"`
	MarketID uint 		`gorm:"index"`
	Time int64			`gorm:"index"`
}

type GdaxOrder struct {
	ID uint 			`gorm:"primary_key"`
	OrderbookID uint 	`gorm:"index"`
	Buy bool
	Rate float64
	//bitfinex supports giving the total number of sell/buyorders.
	//however we should skimp on memory and not add those
	//count int64
	Quantity float64
	Time int64			`gorm:"index"`
}

type GdaxMarket struct {
	ID uint 			`gorm:"primary_key"`
	Ticker string		`gorm:"unique_index:idx_market"`
	Quote string		`gorm:"unique_index:idx_market"`
}

type GdaxTicker struct {
	ID  uint 			`gorm:"primary_key"`
	MarketID uint		`gorm:"index"`
	BestBid float64
	BestAsk float64
	Time int64			`gorm:"index"`
}
