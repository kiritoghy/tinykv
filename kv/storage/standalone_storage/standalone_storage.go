package standalone_storage

import (
	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	// Your Data Here (1).
	conf *config.Config
	db   *badger.DB
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	// Your Code Here (1).
	return &StandAloneStorage{
		conf: conf,
		db:   nil,
	}
}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).
	opts := badger.DefaultOptions
	opts.Dir = s.conf.DBPath
	opts.ValueDir = s.conf.DBPath
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

type StandAloneStorageReader struct {
	txn *badger.Txn
}

func (s *StandAloneStorageReader) GetCF(cf string, key []byte) ([]byte, error) {
	val, err := engine_util.GetCFFromTxn(s.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return val, err
}

func (s *StandAloneStorageReader) IterCF(cf string) engine_util.DBIterator {
	return engine_util.NewCFIterator(cf, s.txn)
}

func (s *StandAloneStorageReader) Close() {
	s.txn.Discard()
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).
	txn := s.db.NewTransaction(false)
	reader := StandAloneStorageReader{
		txn: txn,
	}
	return &reader, nil
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	err := s.db.Update(func(txn *badger.Txn) error {
		for _, m := range batch {
			switch data := m.Data.(type) {
			case storage.Put:
				err := txn.Set(engine_util.KeyWithCF(data.Cf, data.Key), data.Value)
				if err != nil {
					return err
				}
			case storage.Delete:
				err := txn.Delete(engine_util.KeyWithCF(data.Cf, data.Key))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
