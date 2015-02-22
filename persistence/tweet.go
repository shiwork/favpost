package persistence

import (
	"database/sql"
)

type TweetDAO struct {
	*sql.DB
}

func (p *TweetDAO) Prepare() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tweet (
		tweet_id BIGINT UNSIGNED NOT NULL,
		screen_name VARCHAR(100) NOT NULL,
		PRIMARY KEY(tweet_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`

	_, err := p.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (p *TweetDAO) Get(tweet_id int64) *sql.Row {
	sql := `SELECT * FROM tweet WHERE tweet_id = ? LIMIT 1`
	return p.QueryRow(sql, tweet_id)
}

func (p *TweetDAO) Add(tweet_id int64, screenName string) error {
	tx, err := p.Begin()
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO tweet
			(tweet_id, screen_name)
		VALUES
			( ?, ?)
	`
	_, err = tx.Exec(sql, tweet_id, screenName)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (p *TweetDAO) GetByTweet(tweet_id int64, order bool, limit int) (*sql.Rows, error) {
	var sql string
	if order {
		// 指定tweetより新しい物
		sql = `SELECT * FROM tweet WHERE tweet_id > ? LIMIT ?`
	} else {
		// 指定tweetより古いもの
		sql = `SELECT * FROM tweet WHERE tweet_id < ? LIMIT ?`
	}
	return p.Query(sql, tweet_id, limit)
}

func (p *TweetDAO) GetByLimit(limit int) (*sql.Rows, error) {
	sql := `SELECT * FROM tweet ORDER BY tweet_id DESC LIMIT ?`
	return p.Query(sql, limit)
}