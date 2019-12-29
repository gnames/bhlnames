package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func NewDb(host, user, pass, dbname string) (*gorm.DB, error) {
	opts := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable",
		host, user, dbname, pass)
	db, err := gorm.Open("postgres", opts)
	return db, err
}
