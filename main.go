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
	"github.com/getsentry/raven-go"
	"strconv"
	"CIP-exchange-consumer-gdax/pushers"
)

func init(){
	useDotenv := true
	if os.Getenv("PRODUCTION") == "true"{
		useDotenv = false
	}

	// this loads all the constants stored in the .env file (not suitable for production)
	// set variables in supervisor then.
	if useDotenv {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}
	raven.SetDSN(os.Getenv("RAVEN_DSN"))
}

func main() {
	var wsDialer ws.Dialer
	client := gdax.NewClient("", "", "")

	localdb, err := gorm.Open(os.Getenv("DB"), os.Getenv("DB_URL"))
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
	}
	defer localdb.Close()
	err = localdb.AutoMigrate(&db.GdaxOrderBook{}, &db.GdaxOrder{}, &db.GdaxMarket{}, &db.GdaxTicker{}).Error
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
	}

	remotedb, err := gorm.Open(os.Getenv("R_DB"), os.Getenv("R_DB_URL"))
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
	}
	defer remotedb.Close()

	err = localdb.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;").Error
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
	}
	err = localdb.Exec("SELECT create_hypertable('gdax_orders', 'time', 'orderbook_id', if_not_exists => TRUE)").Error
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
	}
	err = localdb.Exec("SELECT create_hypertable('gdax_tickers', 'time', 'market_id', if_not_exists => TRUE)").Error
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
	}
	err = localdb.Exec("SELECT create_hypertable('gdax_order_books', 'time', 'market_id', if_not_exists => TRUE)").Error
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
	}

	//start a replication worker
	limit,  err:= strconv.ParseInt(os.Getenv("REPLICATION_LIMIT"), 10, 64)
	replicator := pushers.Replicator{Local:*localdb, Remote:*remotedb, Limit:limit}
	go replicator.Start()


	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
	}

	Products, err := client.GetProducts()
	if err != nil{
		raven.CaptureErrorAndWait(err, nil)
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
		raven.CaptureErrorAndWait(err, nil)
	}

	message := gdax.Message{}
	// we pass a map to keep track of the different worker channnels to the message
	//consumer
	chanGuide := make(map[string]chan gdax.Message)
	ReadWSConn(wsConn, message, localdb, chanGuide)
	//block the main process
}

func ReadWSConn(conn *ws.Conn, message gdax.Message, gormdb *gorm.DB, chanGuide map[string]chan gdax.Message){
	for true {
		if err := conn.ReadJSON(&message); err != nil {
			log.Panic(err)

		}
		consumer.MessageConsumer(message, *gormdb, 3, chanGuide)
	}
}
