package main

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/gophergala/matching-snuggies/slicerjob"
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
	bucketName := "gCodeFiles"
	return DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b(bucketName))
		if bucket == nil {
			return fmt.Errorf("%v bucket doesn't exist!", bucketName)
		}
		return bucket.Put(b(key), b(value))
	})
}

func PutJob(key string, job *slicerjob.Job) error {
	jsonJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return DB.Update(func(tx *bolt.Tx) error {
		bucketName := "jobs"
		bucket := tx.Bucket(b(bucketName))
		if bucket == nil {
			return fmt.Errorf("%v bucket doesn't exist!", bucketName)
		}
		return bucket.Put(b(key), jsonJob)
	})
}

func ViewGCodeFile(key string) (val string, err error) {
	err = DB.View(func(tx *bolt.Tx) error {
		val = string(tx.Bucket(b("gCodeFiles")).Get(b(key)))
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func ViewJob(key string) (*slicerjob.Job, error) {
	var job = new(slicerjob.Job)
	err := DB.View(func(tx *bolt.Tx) error {
		jsonJob := tx.Bucket(b("jobs")).Get(b(key))
		return json.Unmarshal(jsonJob, job)
	})
	return job, err
}
