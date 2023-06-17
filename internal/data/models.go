package data

import "database/sql"

type UserModel struct {
	DB *sql.DB
}

type MovieModel struct {
	DB *sql.DB
}

type Models struct {
	Movies MovieModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
		Users: UserModel{
			DB: db,
		},
	}
}
