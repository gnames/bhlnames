package db

import (
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/gnames/lib/sys"
)

// InitBadger finds and initializes connection to a badger key-value store.
// If the store does not exist, InitBadger creates it.
func InitKeyVal(dir string) *badger.DB {
	options := badger.DefaultOptions(dir)
	options.Logger = nil
	bdb, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	return bdb
}

func GetValue(kv *badger.DB, key string) int {
	txn := kv.NewTransaction(false)
	defer func() {
		err := txn.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}()
	val, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		// log.Printf("%s not found", key)
		// log.Fatal(err)
		return 0
	} else if err != nil {
		log.Fatal(err)
	}
	var res []byte
	res, err = val.ValueCopy(res)
	if err != nil {
		log.Fatal(err)
	}
	id, _ := strconv.Atoi(string(res))
	return id
}

func ResetKeyVal(dir string) error {
	return sys.CleanDir(dir)
}
