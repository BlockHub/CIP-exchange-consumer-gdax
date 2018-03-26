package main

import(
	ws "github.com/gorilla/websocket"
	"github.com/preichenberger/go-gdax"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"CIP-exchange-consumer-gdax/pkg/consumer"
	"os"
	"github.com/joho/godotenv"
	"log"
	"CIP-exchange-consumer-gdax/internal/db"
)



func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var wsDialer ws.Dialer
	client := gdax.NewClient("", "", "")

	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}

	gormdb, err := gorm.Open(os.Getenv("DB"), os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer gormdb.Close()
	gormdb.AutoMigrate(&db.GdaxOrderBook{}, &db.GdaxOrder{}, &db.GdaxMarket{}, &db.GdaxTicker{})

	err = gormdb.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;").Error
	if err != nil{
		panic(err)
	}
	err = gormdb.Exec("SELECT create_hypertable('gdax_orders', 'time', if_not_exists => TRUE)").Error
	if err != nil{
		panic(err)
	}
	err = gormdb.Exec("SELECT create_hypertable('gdax_tickers', 'time', if_not_exists => TRUE)").Error
	if err != nil{
		panic(err)
	}
	err = gormdb.Exec("SELECT create_hypertable('gdax_order_books', 'time', if_not_exists => TRUE)").Error
	if err != nil{
		panic(err)
	}

	Products, err := client.GetProducts()
	if err != nil{
		panic(err)
	}

	// gdax offers only 8 products, will be a while until they offer 100
	var ProductIds = []string{}
	for _, product := range Products{
		ProductIds = append(ProductIds, product.Id)
	}


	subscribe := gdax.Message{
		Type: "subscribe",
		Channels: []gdax.MessageChannel{
			gdax.MessageChannel{
				Name: "level2",
				ProductIds: ProductIds,
			},
			gdax.MessageChannel{
				Name: "ticker",
				ProductIds: ProductIds,
			},
		},
	}

	if err := wsConn.WriteJSON(subscribe); err != nil {
		println(err.Error())
	}

	message := gdax.Message{}
	// we pass a map to keep track of the different worker channnels to the message
	//consumer
	chanGuide := make(map[string]chan gdax.Message)
	ReadWSConn(wsConn, message, gormdb, chanGuide)
	//block the main process
}

func ReadWSConn(conn *ws.Conn, message gdax.Message, gormdb *gorm.DB, chanGuide map[string]chan gdax.Message){
	for true {
		if err := conn.ReadJSON(&message); err != nil {
			panic(err)

		}
		consumer.MessageConsumer(message, *gormdb, 3, chanGuide)
	}
}
