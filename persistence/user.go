package persistence

import (
	"database/sql"
	"time"
)

type UserPersistence struct {
	*sql.DB
}

func (p UserPersistence) Prepare() error {
	sql := `
	CREATE TABLE IF NOT EXISTS user (
		id	BIGINT UNSIGNED NOT NULL,
		screen_name VARCHAR(100) NOT NULL,
		created TIMESTAMP NOT NULL,
		PRIMARY KEY(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := p.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (p UserPersistence) Get(user_id int64) *sql.Row {
	sql := `SELECT * FROM user WHERE id = ? LIMIT 1`
	return p.QueryRow(sql, user_id)
}

func (p UserPersistence) Add(user_id int64, screenName string) error {
	tx, err := p.Begin()
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO access_token
			(user_id, screen_name, created)
		VALUES
			( ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			screen_name = VALUES(screen_name),
			created = VALUES(created)
	`

	_, err = tx.Exec(sql, user_id, screenName, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (p UserPersistence) GetAll() (*sql.Rows, error) {
	sql := `SELECT * FROM user`
	return p.Query(sql)
}
