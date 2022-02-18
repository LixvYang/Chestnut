// Package storage provides storage for chestnut.
package storage

import (
	"errors"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
)

var (
	DefaultLogFileSize int64 = 16 << 20
	DefaultMemTableSize int64 = 8 << 20
	DefaultMaxEntries uint32 = 50000
	DefaultBlockCacheSize int64 = 32 << 20
	DefaultCompressionType = options.Snappy
	DefaultPrefetchSize = 10
)

type CSBadger struct {
	db *badger.DB
}

func (s *CSBadger) Init(path string) error {
	var err error
	s.db, err = badger.Open(badger.DefaultOptions(path).WithValueLogFileSize(DefaultLogFileSize).WithMemTableSize(DefaultMemTableSize).WithValueLogMaxEntries(DefaultMaxEntries).WithBlockCacheSize(DefaultBlockCacheSize).WithCompression(DefaultCompressionType).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return err
	}
	return nil
}

func (s *CSBadger) Close() error {
	return s.db.Close()
}

func (s *CSBadger) Set(key []byte, val []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val)
		err := txn.SetEntry(e)
		return err
	})
}

func (s *CSBadger) Delete(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		return err
	})
}

func (s *CSBadger) Get(key []byte) ([]byte, error) {
	var val []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	return val, err
}

func (s *CSBadger) IsExist(key []byte) (bool, error) {
	var ret bool

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 1
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Seek(key)
		ret = it.ValidForPrefix(key)
		return nil
	})
	if err == nil {
		return ret, nil
	}
	return false, err
}



func (s *CSBadger) PrefixForeach(prefix []byte, fn func([]byte, []byte, error) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = DefaultPrefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ferr := fn(key, val, nil)
			if ferr != nil {
				return ferr
			}
		}
		return nil
	})
	return err
}

func (s *CSBadger) PrefixForeachKey(prefix []byte, valid []byte, reverse bool, fn func([]byte, error) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 20
		opts.PrefetchValues = false
		opts.Reverse = reverse
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			ferr := fn(key, nil)
			if ferr != nil {
				return ferr
			}
		}
		return nil
	})
	return err
}

func (s *CSBadger) Foreach(fn func([]byte, []byte, error) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = DefaultPrefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ferr := fn(key, val, nil)
			if ferr != nil {
				return ferr
			}
		}
		return nil
	})
	return err
}

func (s *CSBadger) BatchWrite(keys [][]byte, values [][]byte) error {
	if len(keys) != len(values) {
		return errors.New("keys' and values' length should be equal")
	}

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	for i, k := range keys {
		v := values[i]
		e := badger.NewEntry(k, v)
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
	}
	return txn.Commit()

}

func (s *CSBadger) GetSequence(key []byte, bandwidth uint64) (Sequence, error) {
	return s.db.GetSequence(key, bandwidth)
}
