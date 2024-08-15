// Package database
// Author: Evsikov Artem

package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)

const psql = `host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`

func GetConnection() *sqlx.DB {
	psqlInfo := fmt.Sprintf(psql, os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
