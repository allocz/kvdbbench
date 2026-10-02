package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	//"github.com/cockroachdb/pebble/v2"
	"github.com/cockroachdb/pebble/v2"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const sizeTarget = 256 << 20

func gen32(buf []byte, i int) []byte {
	if len(buf) < 32 {
		buf = make([]byte, 32)
	}

	var v [8]byte
	binary.LittleEndian.PutUint64(v[:], uint64(i))
	hash := sha256.Sum256(v[:])
	copy(buf, hash[:])

	return buf
}

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func TestLevelDBBatch(t *testing.T) {
	var buf [32]byte

	hdir, err := os.UserHomeDir()
	must(err)
	dbdir := path.Join(hdir,"tmpldb")
	os.RemoveAll(dbdir)
	must(os.MkdirAll(dbdir, 0o755))
	defer os.RemoveAll(dbdir)

	const totalEntries = sizeTarget / 64

	opts := opt.Options{}
	db, err := leveldb.OpenFile(dbdir, &opts)
	must(err)
	defer db.Close()

	// create N entries
	createStart := time.Now()
	batch := &leveldb.Batch{}
	for i := range totalEntries {
		gen32(buf[:], i)
		batch.Put(buf[:], buf[:])
	}
	err = db.Write(batch, &opt.WriteOptions{Sync: true})
	must(err)
	batch.Reset()
	createElapsed := time.Since(createStart)

	// read N entries
	readStart := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, err := db.Get(buf[:], nil)
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
	}
	err = db.Write(batch, &opt.WriteOptions{Sync: true})
	must(err)
	batch.Reset()
	deleteElapsed := time.Since(deleteStart)

	// read N/2 entries
	read2Start := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, err := db.Get(buf[:], nil)
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
	err = db.Write(batch, &opt.WriteOptions{Sync: true})
	must(err)
	batch.Reset()
	delete2Elapsed := time.Since(delete2Start)

	err = db.Close()
	must(err)
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

func TestPebbleBatch(t *testing.T) {
	var buf [32]byte

	hdir, err := os.UserHomeDir()
	must(err)
	dbdir := path.Join(hdir,"tmpldb")
	os.RemoveAll(dbdir)
	must(os.MkdirAll(dbdir, 0o755))
	defer os.RemoveAll(dbdir)

	const totalEntries = sizeTarget / 64

	dbOpts := pebble.DefaultOptions()
	db, err := pebble.Open(dbdir, dbOpts)
	must(err)

	// create N entries
	createStart := time.Now()
	batch := db.NewBatch()
	for i := range totalEntries {
		gen32(buf[:], i)
		err = batch.Set(buf[:], buf[:], pebble.NoSync)
		must(err)
	}
	err = db.Apply(batch, pebble.Sync)
	must(err)
	batch.Reset()
	createElapsed := time.Since(createStart)

	// read N entries
	readStart := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, c, err := db.Get(buf[:])
		must(err)
		if !bytes.Equal(buf[:], v) {
			t.Fatal("unexpected value")
		}
		err = c.Close()
		must(err)
	}
	readElapsed := time.Since(readStart)

	// delete N/2 entries
	deleteStart := time.Now()
	for i := totalEntries/2; i < totalEntries; i++ {
		gen32(buf[:], i)
		err = batch.Delete(buf[:], pebble.NoSync)
		must(err)
	}
	err = db.Apply(batch, pebble.Sync)
	must(err)
	batch.Reset()
	deleteElapsed := time.Since(deleteStart)

	// read N/2 entries
	read2Start := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		v, c, err := db.Get(buf[:])
		must(err)
		if !bytes.Equal(buf[:], v) {
			t.Fatal("unexpected value")
		}
		err = c.Close()
		must(err)
	}
	read2Elapsed := time.Since(read2Start)

	// delete N/2 entries
	delete2Start := time.Now()
	for i := range totalEntries/2 {
		gen32(buf[:], i)
		err = batch.Delete(buf[:], pebble.NoSync)
		must(err)
	}
	err = db.Apply(batch, pebble.Sync)
	must(err)
	batch.Reset()
	delete2Elapsed := time.Since(delete2Start)

	err = db.Close()
	must(err)
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
