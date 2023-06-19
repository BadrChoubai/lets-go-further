package data

import "database/sql"

type UserModel struct{ DB *sql.DB }
type MovieModel struct{ DB *sql.DB }
type TokenModel struct{ DB *sql.DB }

type Models struct {
	Movies MovieModel
	Users  UserModel
	Token  TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
		Token:  TokenModel{DB: db},
	}
}
