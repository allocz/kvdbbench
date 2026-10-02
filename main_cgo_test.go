//go:build cgo

package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jmhodges/levigo"
)

func TestLeviGoBatch(t *testing.T) {
	var buf [32]byte

	hdir, err := os.UserHomeDir()
	must(err)
	dbdir := path.Join(hdir,"tmpldb")
	os.RemoveAll(dbdir)
	must(os.MkdirAll(dbdir, 0o755))
	defer os.RemoveAll(dbdir)

	const totalEntries = sizeTarget / 64

	dbOpts := levigo.NewOptions()
	dbOpts.SetCreateIfMissing(true)
	db, err := levigo.Open(dbdir, dbOpts)
	must(err)

	// create N entries
	createStart := time.Now()
	batch := levigo.NewWriteBatch()
	for i := range totalEntries {
		gen32(buf[:], i)
		batch.Put(buf[:], buf[:])
	}
	wSync := levigo.NewWriteOptions()
	wSync.SetSync(true)
	err = db.Write(wSync, batch)
	must(err)
	batch.Clear()
	createElapsed := time.Since(createStart)

	// read N entries
	readStart := time.Now()
	ro := levigo.NewReadOptions()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, err := db.Get(ro, buf[:])
		must(err)
		if !bytes.Equal(buf[:], v) {
			t.Fatal("unexpected value")
		}
	}
	readElapsed := time.Since(readStart)

	// delete N/2 entries
	deleteStart := time.Now()
	for i := totalEntries/2; i < totalEntries; i++ {
		gen32(buf[:], i)
		batch.Delete(buf[:])
		must(err)
	}
	err = db.Write(wSync, batch)
	must(err)
	batch.Clear()
	deleteElapsed := time.Since(deleteStart)

	// read N/2 entries
	read2Start := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, err := db.Get(ro, buf[:])
		must(err)
		if !bytes.Equal(buf[:], v) {
			t.Fatal("unexpected value")
		}
	}
	read2Elapsed := time.Since(read2Start)

	// delete N/2 entries
	delete2Start := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		batch.Delete(buf[:])
	}
	err = db.Write(wSync, batch)
	must(err)
	batch.Clear()
	delete2Elapsed := time.Since(delete2Start)

	wSync.Close()
	ro.Close()
	batch.Close()
	db.Close()
	totalElapsed := time.Since(createStart)

	fmt.Printf(
		`
		createN: %fs
		readN: %fs
		deleteN/2: %fs
		readN/2: %fs
		deleteN/2: %fs
		total: %fs
		`,
		createElapsed.Seconds(),
		readElapsed.Seconds(),
		deleteElapsed.Seconds(),
		read2Elapsed.Seconds(),
		delete2Elapsed.Seconds(),
		totalElapsed.Seconds(),
	)
}
