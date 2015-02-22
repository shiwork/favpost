package persistence

import (
	"database/sql"
	"time"
)

type AccessTokenPersistence struct {
	*sql.DB
}

func (p AccessTokenPersistence) Prepare() error {
	sql := `
	CREATE TABLE IF NOT EXISTS access_token (
		user_id	BIGINT UNSIGNED NOT NULL,
		token VARCHAR(100) NOT NULL,
		secret VARCHAR(100) NOT NULL,
		created TIMESTAMP NOT NULL,
		PRIMARY KEY(user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := p.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (p AccessTokenPersistence) Get(user_id int64) *sql.Row {
	sql := `SELECT * FROM access_token WHERE user_id = ?`
	return p.QueryRow(sql, user_id)
}

func (p AccessTokenPersistence) Delete(user_id int64) error {
	tx, err := p.Begin()
	if err != nil {
		return err
	}

	sql := `DELETE FROM access_token WHERE user_id = ?`

	_, err = tx.Exec(sql, user_id)
	if err != nil {
		// error log
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (p AccessTokenPersistence) AddOrUpdate(user_id int64, token string, secret string) error {
	tx, err := p.Begin()
	if err != nil {
		return err
	}

	sql := `
	INSERT INTO access_token
		(user_id, token, secret, created)
	VALUES
		( ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		token = VALUES(token),
		secret = VALUES(secret),
		created = VALUES(created)
	`

	_, err = tx.Exec(sql, user_id, token, secret, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
