package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"flag"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shiwork/favpost/config"
	"github.com/shiwork/favpost/model"
	"github.com/shiwork/favpost/server"
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

	flag.Set("bind", ":9000")

	anaconda.SetConsumerKey(conf.Consumer.ConsumerKey)
	anaconda.SetConsumerSecret(conf.Consumer.ConsumerSecret)

	//var user_id = int64(90649479) // @shiwork
	userRepos := model.GetUserRepository(db)
	go func() {
		for {
			// 酷いけどとりあえず全ユーザーを取得して処理を回す
			users := &[]model.User{}
			users, err = userRepos.GetAll()
			// sleep 5min
			for _, user := range *users {
				//user := &model.User{}
				//user, err = userRepos.Get(user_id)
				api := anaconda.NewTwitterApi(user.AccessToken.Token, user.AccessToken.Secret)
				searchResult, _ := api.GetFavorites(nil)

				for _, tweet := range searchResult {
					if len(tweet.Entities.Media) > 0 {
						ftweet := model.Tweet{
							Id:         tweet.Id,
							ScreenName: tweet.User.ScreenName,
						}

						tweetStore := model.GetTweetRepository(db)
						exists, _ := tweetStore.Exists(ftweet)
						if !exists {
							tweetStore.Add(ftweet)
							model.SlackShare{conf.WebHookURL}.Share(tweet)
						}
					}
				}
			}

			// sleep 5min
			time.Sleep(5 * time.Minute)
		}
	}()

	server.Run(conf, db)
}
