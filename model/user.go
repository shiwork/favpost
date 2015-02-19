package model

import (
	"database/sql"
	"github.com/shiwork/favpost/persistence"
	"time"
)

type AccessToken struct {
	Id         int64
	Token      string
	Secret     string
	Expiration time.Time
	Expired    bool
	Created    time.Time
}

type User struct {
	Id          int64
	ScreenName  string
	AccessToken AccessToken
	Created     time.Time
}

type UserRepository struct {
	userPers  persistence.UserPersistence
	tokenPers persistence.AccessTokenPersistence
}

func GetUserRepository(db *sql.DB) UserRepository {
	return UserRepository{
		userPers:  persistence.UserPersistence{db},
		tokenPers: persistence.AccessTokenPersistence{db},
	}
}

func (r UserRepository) Get(user_id int64) (*User, error) {
	row := r.userPers.Get(user_id)

	user := &User{}
	err := row.Scan(
		&(user.Id),
		&(user.ScreenName),
		&(user.Created),
	)
	if err != nil {
		return nil, err
	}

	row = r.tokenPers.Get(user_id)
	// todo tokenが削除されている場合を考慮する

	token := &AccessToken{}

	err = row.Scan(
		&(token.Id),
		&(token.Token),
		&(token.Secret),
		&(token.Created),
	)
	if err != nil {
		// Tokenがエラーでもとりあえず動かすが、念の為初期化
		token = &AccessToken{}
	}

	user.AccessToken = *token
	return user, nil
}

func (r UserRepository) Exists(Id int64) bool {
	_, err := r.Get(Id)
	switch {
	case err == sql.ErrNoRows:
		return false
	default:
		return true
	}
}

func (r UserRepository) Add(user User) error {
	return r.userPers.Add(user)
}

func (r UserRepository) SaveToken(user User) error {
	return r.tokenPers.AddOrUpdate(user.Id, user.AccessToken)
}
