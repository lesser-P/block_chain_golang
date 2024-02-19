package database

import (
	"testing"
)

func TestPut(t *testing.T) {
	db := New()

	db.Put([]byte("key"), []byte("value"), BlockBucket)
}
