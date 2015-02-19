package persistence
import (
	"database/sql"
	"github.com/shiwork/favpost/model"
)


type UserPersistence struct {
	*sql.DB
}

func (p UserPersistence) Prepare() error {
	sql := `
	CREATE TABLE IF NOT EXISTS user (
		id	BIGINT UNSIGNED NOT NULL,
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

func (p UserPersistence) Get(user_id int64) sql.Row {
	sql := `SELECT * FROM tweets WHERE id = ? LIMIT 1`
	return p.QueryRow(sql, user_id)
}

func (p UserPersistence) Add(user model.User) error {
	return nil
}
