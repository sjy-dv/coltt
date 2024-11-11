// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package bbolt

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/sjy-dv/nnv/backup/store"
	"github.com/sjy-dv/nnv/backup/util"
	"go.etcd.io/bbolt"
)

type boltStore struct {
	db *bbolt.DB
}

const (
	dbFileName = "data.db"
	rootBucket = "root"
)

func Open(dir string) (store.Store, error) {
	dirExists, err := util.PathExists(dir)
	if err != nil {
		return nil, err
	}
	if !dirExists {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, err
		}
	}
	db, err := bbolt.Open(filepath.Join(dir, dbFileName), 0600, nil)
	if err != nil {
		return nil, err
	}
	dataStore := &boltStore{db: db}
	err = dataStore.createRootBucketIfNotExists()
	return dataStore, err
}

func (store *boltStore) createRootBucketIfNotExists() error {
	tx, err := store.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte(rootBucket))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (store *boltStore) Begin(update bool) (store.Tx, error) {
	tx, err := store.db.Begin(update)
	return &boltTx{Tx: tx}, err
}

func (store *boltStore) Close() error {
	return store.db.Close()
}

type boltTx struct {
	*bbolt.Tx
}

func (tx *boltTx) bucket() *bbolt.Bucket {
	return tx.Bucket([]byte(rootBucket))
}

func (tx *boltTx) Set(key, value []byte) error {
	bucket := tx.bucket()
	return bucket.Put(key, value)
}

func (tx *boltTx) Get(key []byte) ([]byte, error) {
	bucket := tx.bucket()
	return bucket.Get(key), nil
}

func (tx *boltTx) Delete(key []byte) error {
	bucket := tx.bucket()
	return bucket.Delete(key)
}

func (tx *boltTx) Cursor(forward bool) (store.Cursor, error) {
	bucket := tx.bucket()
	cursor := bucket.Cursor()
	return &boltCursor{
		Cursor:  cursor,
		forward: forward,
	}, nil
}

func (tx *boltTx) Commit() error {
	return tx.Tx.Commit()
}

func (tx *boltTx) Rollback() error {
	return tx.Tx.Rollback()
}

type boltCursor struct {
	*bbolt.Cursor
	forward bool

	currItem *store.Item
}

func (c *boltCursor) Seek(seek []byte) error {
	key, value := c.Cursor.Seek(seek)
	if key != nil && value != nil {
		c.currItem = &store.Item{
			Key:   key,
			Value: value,
		}
	}

	c.adjustSeek(key, seek)
	return nil
}

func (c *boltCursor) adjustSeek(key []byte, seek []byte) {
	if key != nil && !bytes.Equal(key, seek) && !c.forward {
		key, value := c.Cursor.Prev()
		c.currItem = &store.Item{
			Key:   key,
			Value: value,
		}
	}
}

func (c *boltCursor) Next() {
	var key, value []byte
	if c.forward {
		key, value = c.Cursor.Next()
	} else {
		key, value = c.Cursor.Prev()
	}

	c.currItem = &store.Item{
		Key:   key,
		Value: value,
	}
}

func (c *boltCursor) Valid() bool {
	return c.currItem != nil && c.currItem.Key != nil && c.currItem.Value != nil
}

func (c *boltCursor) Item() (store.Item, error) {
	item := c.currItem
	return *item, nil
}

func (c *boltCursor) Close() error {
	return nil
}
