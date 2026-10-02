# Go key/value db benchmark

* leveldb: `github.com/syndtr/goleveldb/leveldb`
* levigo: `github.com/jmhodges/levigo`
* pebble: `github.com/cockroachdb/pebble`

| db      | cgo | createN | readN  | deleteN/2 | readN/2 | deleteN/2 | total  | max_mem   | tot_read | tot_write |
|---------|-----|---------|--------|-----------|---------|-----------|--------|-----------|----------|-----------|
| leveldb | 0   | 41.16s  | 13.40s | 7.12s     | 13.30s  | 5.11s     | 80.66s | 1.80GiB   | 9.99GiB  | 693.83MiB |
| levigo  | 1   | 13.21s  | 7.88s  | 5.18s     | 15.19s  | 5.16s     | 48.21s | 808.26MiB | 16.29KiB | 747.83MiB |
| pebble  | 0   | 5.01s   | 10.24s | 2.01s     | 18.26s  | 2.48s     | 40.07s | 1.32GiB   | 6.83GiB  | 858.32MiB |
| pebble  | 1   | 4.76s   | 9.72s  | 1.89s     | 16.62s  | 1.86s     | 36.34s | 1.33GiB   | 7.14GiB  | 835.25MiB |
