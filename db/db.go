package database

import (
	"github.com/boltdb/bolt"
	"log"
)

// Open the DB in read-only mode
func OpenRead() *bolt.DB {
	return openDB(true)
}


//Open the DB in write mode
func OpenWrite() *bolt.DB {
	return openDB(false)
}

//Actually OpenDB
func openDB(readOnly bool) *bolt.DB {
	db, err := bolt.Open("./blocks.db", 0600, &bolt.Options{ReadOnly: readOnly})
	if err != nil {
		log.Fatal(err)
	}
	return db
}



