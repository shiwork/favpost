package favpost

import (
	"github.com/ChimeraCoder/anaconda"
	"net/url"
)

func CrawlTwitterFavorite(string access_token, string access_token_secret){
	api := anaconda.NewTwitterApi(access_token, access_token_secret)
	searchResult, _ := api.GetFavorites(url.Values{"count": 200})

	for _, tweet := range searchResult {
		if len(tweet.Entities.Media) > 0 {
		}
	}
}
