package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/gophergala/matching-snuggies/slicerjob"
)

func b(s string) []byte {
	return []byte(s)
}

var DB *bolt.DB

func loadDB(path string) *bolt.DB {
	db, err := bolt.Open(path, 0666, nil)
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

func CancelJob(id string) error {
	log.Println("CancelJob called")
	bucket := "jobs"
	err := DB.View(func(tx *bolt.Tx) error {
		var job = new(slicerjob.Job)
		jsonJob := tx.Bucket(b(bucket)).Get(b(id))
		err := json.Unmarshal(jsonJob, job)
		if err != nil {
			return err
		}
		job.Status = slicerjob.Cancelled
		return PutJob(id, job)
	})
	log.Println("CancelJob end")
	return err
}

func DeleteJob(id string) error {
	bucket := "jobs"
	err := DB.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(b(bucket)).Delete(b(id))
		return err
	})
	return err
}

func DeleteGCodeFile(id string) error {
	bucket := "gCodeFiles"
	err := DB.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(b(bucket)).Delete(b(id))
		return err
	})
	return err
}
