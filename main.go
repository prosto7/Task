package main

//go:generate easyjson -pkg model
import (
	"Task/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func main() {

	var db model.DB

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	db.ConnectDB(viper.GetString("DB.dsn"))

	wg := new(sync.WaitGroup)
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workersCount := viper.GetInt("worker.count")

	err = db.Db.AutoMigrate(&model.Price{})
	if err != nil {
		log.Fatalln(err)
	}

	workers := make([]chan *model.Data, workersCount)

	for {

		resp, err := http.Get("https://api.binance.com/api/v3/ticker/price")
		if err != nil {
			log.Fatalln(err)
		}
		timeStamp := time.Now()
		var p []model.Price

		err = json.NewDecoder(resp.Body).Decode(&p)

		if err != nil {
			log.Fatalln(err)
		}

		for i := 0; i < workersCount; i++ {
			wg.Add(1)
			workers[i] = make(chan *model.Data)
			go worker(ctx, workers[i], db.Db, wg)
		}

		rem := len(p) % workersCount
		part := len(p) / workersCount
		start := 0
		end := rem + part - 1

		for i := 0; i < workersCount && end < len(p); i++ {
			log.Println(start, end)
			workers[i] <- &model.Data{TimeStamp: timeStamp, Prices: p[start:end]}
			start += part
			end += part
			if part == 0 {
				break
			}
		}

		time.Sleep(time.Second * time.Duration(viper.GetInt("worker.frequency")))
	}
}
func worker(ctx context.Context, prices chan *model.Data, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-prices:
			_, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel() // releases resources if slowOperation completes before timeout elapses

			for _, v := range p.Prices {
				tx := db.Model(model.Price{}).Where("symbol = ?", v.Symbol).Updates(model.Price{Symbol: v.Symbol, Price: v.Price, TimeStamp: p.TimeStamp})
				if tx.RowsAffected == 0 {
					v.TimeStamp = time.Now()
					db.Create(&v)
				}
			}
		}
	}
}
