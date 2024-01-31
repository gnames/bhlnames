package db

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/gnsys"
)

// InitBadger finds and initializes connection to a badger key-value store.
// If the store does not exist, InitBadger creates it.
func InitKeyVal(dir string) (*badger.DB, error) {
	options := badger.DefaultOptions(dir)
	options.Logger = nil
	bdb, err := badger.Open(options)
	if err != nil {
		err = fmt.Errorf("db.InitKeyVal: %w", err)
		slog.Error("Cannot open Badger DB", "error", err)
		return nil, err
	}
	return bdb, nil
}

func GetValue(kv *badger.DB, key string) (int, error) {
	txn := kv.NewTransaction(false)
	val, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		return 0, nil
	} else if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		slog.Error("Cannot get value", "error", err)
		return 0, err
	}
	var res []byte
	res, err = val.ValueCopy(res)
	if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		slog.Error("Cannot copy value", "error", err)
		return 0, err
	}
	id, err := strconv.Atoi(string(res))
	if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		slog.Error("Cannot convert value to int", "error", err)
		return 0, err
	}
	err = txn.Commit()
	if err != nil {
		err = fmt.Errorf("db.GetValue: %w", err)
		slog.Error("Cannot commit transaction", "error", err)
		return id, err
	}

	return id, nil
}

func GetValues(kv *badger.DB, keys []string) (map[string][]byte, error) {
	res := make(map[string][]byte)
	txn := kv.NewTransaction(false)
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
	err := txn.Commit()
	if err != nil {
		err = fmt.Errorf("db.GetValues: %w", err)
		slog.Error("Cannot commit transaction", "error", err)
		return nil, err
	}
	return res, nil
}

func ResetKeyVal(dir string) error {
	return gnsys.CleanDir(dir)
}
