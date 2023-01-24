package item

import "github.com/gnames/bhlnames/internal/io/db"

type Item struct {
	db.Item
	db.ItemStats
}
