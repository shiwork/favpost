package favpost

import (
	"database/sql"
	"strconv"
	"github.com/ChimeraCoder/anaconda"
)

type Tweet struct {
	Id         int64
	ScreenName string
}

func (m *Tweet) URL() string {
	return "http://twitter.com/" + m.ScreenName + "/status/" + strconv.FormatInt(m.Id, 10)
}

type TweetRepository struct {
	db *sql.DB
}

func NewTweetRepository(db *sql.DB) *TweetRepository {
	return &TweetRepository{db}
}

func (r *TweetRepository) Add(tweet *anaconda.Tweet) error {
	query := `INSERT INTO tweet (tweet_id, screen_name) VALUES (?, ?)`
	_, err := r.db.Exec(query, tweet.Id, tweet.User.ScreenName)
	if err != nil {
		return err
	}
	return nil
}

func (r *TweetRepository) Exists(tweet *anaconda.Tweet) (bool, error) {
	query := `SELECT * FROM tweet WHERE tweet_id = ? LIMIT 1`
	row := r.db.QueryRow(query, tweet.Id)
	storeTweet := &Tweet{}
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

func (r *TweetRepository) Find(limit int) (*[]Tweet, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 0 {
		limit = 1
	}

	query := `SELECT * FROM tweet ORDER BY tweet_id DESC LIMIT ?`
	rows, err := r.db.Query(query, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []Tweet
	for rows.Next() {
		tweet := &Tweet{}
		err := rows.Scan(
			&(tweet.Id),
			&(tweet.ScreenName),
		)
		if err != nil {
			continue
		}

		tweets = append(tweets, *tweet)
	}

	return &tweets, nil
}