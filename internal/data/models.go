package data

import "database/sql"

type MovieModel struct{ DB *sql.DB }
type PermissionModel struct{ DB *sql.DB }
type TokenModel struct{ DB *sql.DB }
type UserModel struct{ DB *sql.DB }

type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Token       TokenModel
	Users       UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
		Token:  TokenModel{DB: db},
	}
}
