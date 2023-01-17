package db

import (
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

// InitBadger finds and initializes connection to a badger key-value store.
// If the store does not exist, InitBadger creates it.
func InitKeyVal(dir string) *badger.DB {
	options := badger.DefaultOptions(dir)
	options.Logger = nil
	bdb, err := badger.Open(options)
	if err != nil {
		err = fmt.Errorf("db.InitKeyVal: %w", err)
		log.Fatal().Err(err).Msg("")
	}
	return bdb
}

func GetValue(kv *badger.DB, key string) int {
	txn := kv.NewTransaction(false)
	defer func() {
		err := txn.Commit()
		if err != nil {
			err = fmt.Errorf("db.GetValue: %w", err)
			log.Fatal().Err(err).Msg("")
		}
	}()
	val, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		return 0
	} else if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		log.Fatal().Err(err).Msg("")
	}
	var res []byte
	res, err = val.ValueCopy(res)
	if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		log.Fatal().Err(err).Msg("")
	}
	id, _ := strconv.Atoi(string(res))
	return id
}

func GetValues(kv *badger.DB, keys []string) (map[string][]byte, error) {
	res := make(map[string][]byte)
	txn := kv.NewTransaction(false)
	defer func() {
		err := txn.Commit()
		if err != nil {
			err = fmt.Errorf("db.GetValues: %w", err)
			log.Fatal().Err(err).Msg("")
		}
	}()
	for i := range keys {
		val, err := txn.Get([]byte(keys[i]))
		if err == badger.ErrKeyNotFound {
			res[keys[i]] = nil
			continue
		} else if err != nil {
			return res, err
		}
		var bs []byte
		bs, err = val.ValueCopy(bs)
		if err != nil {
			return res, err
		}
		res[keys[i]] = bs
	}
	return res, nil
}

func ResetKeyVal(dir string) error {
	return gnsys.CleanDir(dir)
}