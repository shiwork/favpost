package main

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/shiwork/favpost/config"
	"os"
	"log"
	"database/sql"
	"github.com/shiwork/favpost/storage"
	"time"
	"github.com/shiwork/favpost/model"
	_ "github.com/go-sql-driver/mysql"
)

var confPath = os.Getenv("FAVPOST_CONFIG")

func main() {
	conf, err := config.Parse(confPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", conf.DbDsn)
	if err != nil {
		log.Fatalf("Db initialize error %v\n", err)
	}
	err = storage.Prepare(db)
	if err != nil {
		log.Fatalf("Db prepare error: %v\n", err)
	}

	anaconda.SetConsumerKey(conf.Consumer.ConsumerKey)
	anaconda.SetConsumerSecret(conf.Consumer.ConsumerSecret)
	api := anaconda.NewTwitterApi(conf.AccessToken.Token, conf.AccessToken.Secret)

	for {

		searchResult, _ := api.GetFavorites(nil)

		for _, tweet := range searchResult {
			if len(tweet.Entities.Media) > 0 {

				exists,_ := storage.Exists(db, tweet)
				if !exists {
					storage.Add(db, tweet)
					model.SlackShare{conf.WebHookURL}.Share(tweet)
				}
			}
		}

		// sleep 5min
		time.Sleep(5 * time.Minute)
	}
}
