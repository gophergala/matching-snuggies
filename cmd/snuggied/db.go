package main

import (
	"fmt"

	"github.com/boltdb/bolt"
)

func b(s string) []byte {
	return []byte(s)
}

var DB = loadDB()

func loadDB() *bolt.DB {
	db, err := bolt.Open("db", 0666, nil)
	if err != nil {
		panic(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(b("gCodeFiles"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(b("jobs"))
		if err != nil {
			return err
		}
		return nil
	})
	return db
}

func PutGCodeFile(key string, value string) error {
	return puttIt(key, b(value), "gCodeFiles")
}

func PutJob(key string, value []byte) error {
	return puttIt(key, value, "jobs")
}

func puttIt(key string, value []byte, bucketName string) error {
	return DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b(bucketName))
		if bucket == nil {
			return fmt.Errorf("%v bucket doesn't exist!", bucketName)
		}
		err := bucket.Put(b(key), value)
		if err != nil {
			return err
		}
		return nil
	})
}

func ViewGCodeFile(key string, value string) (string, error) {
	val, err := viewIt(key, "gCodeFiles")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func ViewJob(key string) (val []byte, err error) {
	return viewIt(key, "jobs")
}

func viewIt(key, bucketName string) (val []byte, err error) {
	err = DB.View(func(tx *bolt.Tx) error {
		val = tx.Bucket(b(bucketName)).Get(b(key))
		return nil
	})
	return val, err
}
