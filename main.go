package main

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/shiwork/slack"
	"github.com/shiwork/favpost/config"
	"os"
	"log"
	"database/sql"
	"github.com/shiwork/favpost/storage"
	"time"
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
	incoming := slack.Incoming{WebHookURL: conf.WebHookURL}

	for {

		searchResult, _ := api.GetFavorites(nil)

		for _, tweet := range searchResult {
			if len(tweet.Entities.Media) > 0 {

				exists,_ := storage.Exists(db, tweet)
				if !exists {
					storage.Add(db, tweet)

					incoming.Post(
					slack.Payload{
						Attachments: []slack.Attachment{
							slack.Attachment{
								Pretext: "http://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr,
								Title: tweet.User.Name + " @" + tweet.User.ScreenName,
								TitleLink: tweet.User.URL,
								Text: tweet.Text,
								ImageUrl: tweet.Entities.Media[0].Media_url,
							},
						},
					},
					)
				}
			}
		}

		// sleep 5min
		time.Sleep(5 * time.Minute)
	}
}
