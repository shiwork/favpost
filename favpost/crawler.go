package favpost

import (
	"github.com/ChimeraCoder/anaconda"
	"net/url"
)

func CrawlTwitterFavorite(access_token string, access_token_secret string){
	api := anaconda.NewTwitterApi(access_token, access_token_secret)
	var values []string
	values[0] = "200"
	searchResult, _ := api.GetFavorites(url.Values{"count": values})

	for _, tweet := range searchResult {
		if len(tweet.Entities.Media) > 0 {
		}
	}
}
