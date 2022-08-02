package sqlite3

import (
	"fmt"
	"log"
	"strconv"
	"testing"
)

func MemoryTable(rows int) (*Conn, error) {
	conn, err := Open(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %s", err)
	}

	if err := conn.Exec(`CREATE TABLE t1 (c1 TEXT, c2 TEXT, c3 INTEGER);`); err != nil {
		return nil, fmt.Errorf("could not create table: %s", err)
	}

	// insert some data
	for i := 0; i < rows; i++ {
		if err := conn.Exec(`INSERT INTO t1 (c1, c2, c3) VALUES (?, ?, ?);`,
			"foo"+strconv.Itoa(i),
			"bar"+strconv.Itoa(i), i); err != nil {
			return nil, fmt.Errorf("could not insert into table: %s", err)
		}
	}

	return conn, nil
}

func TestConn_ExecWithCallback(t *testing.T) {
	type args struct {
		sql      string
		callback ExecCallback
		data     interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test callback",
			args: args{
				sql: "select * from t1 limit 200",
				callback: func(data, columns []string) int {
					t.Logf("data: %v", data)
					t.Logf("cols: %v", columns)
					// 0 means success
					return 0
				},
			},
			wantErr: false,
		},
		{
			name: "test callback with negative return",
			args: args{
				sql: "select * from t1 limit 200",
				callback: func(data, columns []string) int {
					t.Logf("data: %v", data)
					t.Logf("cols: %v", columns)
					// -1 returns "sqlite3: query aborted [4]"
					return -1
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := MemoryTable(1)
			if err != nil {
				t.Fatalf("cannot open conn: %s", err)
			}
			defer c.Close()

			if err := c.ExecWithCallback(tt.args.sql, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("Conn.ExecWithCallback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type Table struct {
	count int
}

func (t *Table) Callback(data, names []string) int {
	t.count++
	return 0
}

func (t *Table) DataCallback(data []string) int {
	t.count++
	return 0
}

func Benchmark_ExecWithCallback(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
		table *Table
	}{
		{
			name:  "1 row",
			count: 1,
			table: &Table{},
		},
		{
			name:  "10 row",
			count: 10,
			table: &Table{},
		},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			c, err := MemoryTable(bb.count)
			if err != nil {
				b.Fatalf("cannot open conn: %s", err)
			}
			defer c.Close()

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_ = c.ExecWithCallback("select * from t1", bb.table.Callback)
			}

			b.Logf("tableCount: %d", bb.table.count)
		})
	}
}

func Benchmark_ExecWithDataCallback(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
		table *Table
	}{
		{
			name:  "1 row",
			count: 1,
			table: &Table{},
		},
		{
			name:  "10 row",
			count: 10,
			table: &Table{},
		},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			c, err := MemoryTable(bb.count)
			if err != nil {
				b.Fatalf("cannot open conn: %s", err)
			}
			defer c.Close()

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_ = c.ExecWithDataCallback("select * from t1", bb.table.DataCallback)
			}

			b.Logf("tableCount: %d", bb.table.count)
		})
	}
}

func ExampleConn_ExecWithCallback() {
	conn, _ := Open(":memory:")

	if err := conn.Exec(`CREATE TABLE t1 (c1 TEXT, c2 TEXT, c3 INTEGER);`); err != nil {
		log.Fatalf("could not create table: %s", err)
	}

	if err := conn.Exec(`INSERT INTO t1 (c1, c2, c3) VALUES (?, ?, ?);`, "foo", "bar", 1); err != nil {
		log.Fatalf("could not insert into table: %s", err)
	}

	conn.ExecWithCallback("select * from t1", func(columnData, columnNames []string) int {
		for i := range columnNames {
			fmt.Printf("%s: %s\n", columnNames[i], columnData[i])
		}

		return 0
	})

	// Output: c1: foo
	// c2: bar
	// c3: 1
}

func ExampleConn_ExecWithDataCallback() {
	conn, _ := Open(":memory:")

	if err := conn.Exec(`CREATE TABLE t1 (c1 TEXT, c2 TEXT, c3 INTEGER);`); err != nil {
		log.Fatalf("could not create table: %s", err)
	}

	if err := conn.Exec(`INSERT INTO t1 (c1, c2, c3) VALUES (?, ?, ?);`, "foo", "bar", 1); err != nil {
		log.Fatalf("could not insert into table: %s", err)
	}

	conn.ExecWithDataCallback("select * from t1", func(columnData []string) int {
		for i := range columnData {
			fmt.Printf("col[%d]: %s\n", i, columnData[i])
		}

		return 0
	})

	// Output: col[0]: foo
	// col[1]: bar
	// col[2]: 1
}
