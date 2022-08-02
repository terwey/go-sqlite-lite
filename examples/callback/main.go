package main

import (
	"log"
	"strconv"

	"github.com/productsupcom/go-sqlite-lite/sqlite3"
)

func main() {
	conn, err := sqlite3.Open(":memory:")
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	defer conn.Close()

	if err := conn.Exec(`CREATE TABLE t1 (c1 TEXT, c2 TEXT, c3 INTEGER);`); err != nil {
		log.Fatalf("could not create table: %s", err)
	}

	// insert some data
	for i := 0; i < 100; i++ {
		if err := conn.Exec(`INSERT INTO t1 (c1, c2, c3) VALUES (?, ?, ?);`,
			"foo"+strconv.Itoa(i),
			"bar"+strconv.Itoa(i), i); err != nil {
			log.Fatalf("could not insert into table: %s", err)
		}
	}

	err = conn.ExecWithCallback("select * from t1 limit 1", func(data, names []string) int {
		log.Printf("data: %s", data)
		log.Printf("names: %s", names)
		return 0
	})
	if err != nil {
		log.Fatalf("could not call ExecWithCallback: %s", err)
	}

	err = conn.ExecWithDataCallback("select * from t1 limit 1", func(data []string) int {
		log.Printf("data only: %s", data)
		return 0
	})
	if err != nil {
		log.Fatalf("could not call ExecWithCallback: %s", err)
	}

	// call it with a struct that has a func
	table := &Table{}
	err = conn.ExecWithCallback("select * from t1", table.Callback)
	if err != nil {
		log.Fatalf("could not call ExecWithCallback: %s", err)
	}

	log.Printf("table rows: %d", table.count)
}

type Table struct {
	count int
}

func (t *Table) Callback(data, names []string) int {
	t.count++
	return 0
}
