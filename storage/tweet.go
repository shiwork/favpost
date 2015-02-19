package storage

import (
	"database/sql"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"time"
)

type Tweet struct {
	Id int64
}

func Prepare(db *sql.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS tweets (
		id	BIGINT UNSIGNED NOT NULL,
		created TIMESTAMP NOT NULL,
		PRIMARY KEY(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func Add(db *sql.DB, tweet anaconda.Tweet) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	sql := `
		INSERT INTO tweets
		(id, created)
		VALUES
		( ?, ?)
	`
	_, err = tx.Exec(sql, tweet.Id, time.Now())
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// tweetが送信済みかどうかをチェックする
func Exists(db *sql.DB, tweet anaconda.Tweet) (bool, error) {
	existsSql := `SELECT * FROM tweets WHERE id = ? LIMIT 1`
	storeTweet := &Tweet{}
	err := db.QueryRow(existsSql, tweet.Id).Scan(
		&(storeTweet.Id),
	)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return true, err
	default:
		return true, nil
	}
}
