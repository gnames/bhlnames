package dbio

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/gnsys"
)

// InitBadger finds and initializes connection to a badger key-value store.
// If the store does not exist, InitBadger creates it.
func InitKeyVal(dir string, readonly bool) (*badger.DB, error) {
	options := badger.DefaultOptions(dir)
	options.Logger = nil
	options.ReadOnly = readonly
	bdb, err := badger.Open(options)
	if err != nil {
		err = fmt.Errorf("db.InitKeyVal: %w", err)
		slog.Error("Cannot open Badger DB.", "error", err)
		return nil, err
	}
	return bdb, nil
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
		slog.Error("Cannot commit transaction.", "error", err)
		return nil, err
	}
	return res, nil
}

func ResetKeyVal(dir string) error {
	stat := gnsys.GetDirState(dir)
	if stat == gnsys.DirAbsent {
		return os.MkdirAll(dir, 0755)
	}
	return gnsys.CleanDir(dir)
}
