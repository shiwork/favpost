package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"flag"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kavu/go-resque"
	_ "github.com/kavu/go-resque/godis"
	"../../favpost"
	"github.com/simonz05/godis/redis"
	"fmt"
	"strconv"
)

var confPath = os.Getenv("FAVPOST_CONFIG")

func main() {
	confPath := flag.String("c", "", "configuration file path")
	flag.Parse()

	conf, err := favpost.Parse(*confPath)
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
	userRepos := favpost.NewUserRepository(db)

	// enqueue
	redisClient := redis.New("tcp:127.0.0.1:6379", 0, "")
	botqueue := resque.NewRedisEnqueuer("godis", redisClient)

	go func() {
		for {
			// 酷いけどとりあえず全ユーザーを取得して処理を回す
			users := &[]favpost.User{}
			users, err = userRepos.GetAll()
			// sleep 5min
			for _, user := range *users {
				//user := &model.User{}
				//user, err = userRepos.Get(user_id)
				api := anaconda.NewTwitterApi(user.AccessToken.Token, user.AccessToken.Secret)
				searchResult, _ := api.GetFavorites(nil)

				for _, tweet := range searchResult {
					if len(tweet.Entities.Media) > 0 {
						tweetStore := favpost.NewTweetRepository(db)
						exists, _ := tweetStore.Exists(tweet)
						if !exists {
							tweetStore.Add(tweet)
							favpost.SlackShare{conf.WebHookURL}.Share(tweet)

							// bot enqueue
							_, err := botqueue.Enqueue("resque:queue:favpostbot", "Favpost", strconv.FormatInt(tweet.Id, 10), tweet.User.ScreenName)
							if err != nil {
								fmt.Println(err)
							}

						}

					}
				}
			}

			// sleep 5min
			time.Sleep(5 * time.Minute)
		}
	}()

	favpost.Run(conf, db)
}
