package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shiwork/favpost/config"
	"github.com/shiwork/favpost/model"
	"github.com/shiwork/favpost/server"
	"github.com/shiwork/favpost/storage"
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

	// このへんは後で
	//	var callbackURL = "http://localhost:9001/auth/callback"
	//	credentials, err := anaconda.AuthorizationURL(callbackURL)

	//var user_id = int64(90649479) // @shiwork
	userRepos := model.GetUserRepository(db)
	go func() {
		for {
			// 酷いけどとりあえず全ユーザーを取得して処理を回す
			users := &[]model.User{}
			users, err = userRepos.GetAll()
			// sleep 5min
			time.Sleep(5 * time.Minute)

			for _, user := range *users {
				//user := &model.User{}
				//user, err = userRepos.Get(user_id)
				api := anaconda.NewTwitterApi(user.AccessToken.Token, user.AccessToken.Secret)
				searchResult, _ := api.GetFavorites(nil)

				for _, tweet := range searchResult {
					if len(tweet.Entities.Media) > 0 {

						exists, _ := storage.Exists(db, tweet)
						if !exists {
							storage.Add(db, tweet)
							model.SlackShare{conf.WebHookURL}.Share(tweet)
						}
					}
				}
			}

			// sleep 5min
			time.Sleep(5 * time.Minute)
		}
	}()

	/*
		fmt.Println("run http")
		http.HandleFunc("/", server.SayhelloName)
		err = http.ListenAndServe(":9090", nil)
		if err != nil {
			fmt.Println("http Listen error")
			log.Fatal("ListenAndServe: ", err)
		}
		fmt.Println("end of main func.")
	*/
	server.Run(conf, db)
}
