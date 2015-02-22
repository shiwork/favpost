package model

import (
	"database/sql"

	"strconv"

	"github.com/shiwork/favpost/persistence"
)

type Tweet struct {
	Id         int64
	ScreenName string
}

func (m *Tweet) URL() string {
	return "http://twitter.com/" + m.ScreenName + "/status/" + strconv.FormatInt(m.Id, 10)
}

type TweetRepository struct {
	tweetDAO persistence.TweetDAO
}

func GetTweetRepository(db *sql.DB) TweetRepository {
	tweetDAO := persistence.TweetDAO{db}
	return TweetRepository{tweetDAO}
}

func (r *TweetRepository) Add(tweet Tweet) error {
	return r.tweetDAO.Add(tweet.Id, tweet.ScreenName)
}

func (r *TweetRepository) Exists(tweet Tweet) (bool, error) {
	storeTweet := &Tweet{}
	row := r.tweetDAO.Get(tweet.Id)
	err := row.Scan(
		&(storeTweet.Id),
		&(storeTweet.ScreenName),
	)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (r *TweetRepository) FindByOrder(tweet Tweet, order bool, limit int) (*[]Tweet, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 0 {
		limit = 1
	}
	rows, err := r.tweetDAO.GetByTweet(tweet.Id, order, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []Tweet
	for rows.Next() {
		gtweet := &Tweet{}
		err := rows.Scan(
			&(gtweet.Id),
			&(gtweet.ScreenName),
		)
		if err != nil {
			continue
		}

		tweets = append(tweets, *gtweet)
	}

	return &tweets, nil
}
