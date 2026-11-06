package sqlite3

import (
	"reflect"
	"testing"
)

// TestBindMultibytesPadding inserts rows through one prepared statement, as a
// bulk loader does, so a column left unbound would keep the previous row's value.
func TestBindMultibytesPadding(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want [][]string
	}{
		{
			name: "full rows",
			rows: [][]string{{"1", "2", "3"}, {"4", "5", "6"}},
			want: [][]string{{"1", "2", "3"}, {"4", "5", "6"}},
		},
		{
			name: "short row after a full row",
			rows: [][]string{{"1", "2", "3"}, {"4", "5"}},
			want: [][]string{{"1", "2", "3"}, {"4", "5", ""}},
		},
		{
			name: "single value after a full row",
			rows: [][]string{{"6", "7", "8"}, {"9"}},
			want: [][]string{{"6", "7", "8"}, {"9", "", ""}},
		},
		{
			name: "empty values",
			rows: [][]string{{"1", "", "3"}, {"", "", ""}},
			want: [][]string{{"1", "", "3"}, {"", "", ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := Open(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			if err := conn.Exec("CREATE TABLE x(a TEXT, b TEXT, c TEXT)"); err != nil {
				t.Fatal(err)
			}

			insert, err := conn.Prepare("INSERT INTO x (a, b, c) VALUES (?, ?, ?)")
			if err != nil {
				t.Fatal(err)
			}
			defer insert.Close()

			for _, row := range tt.rows {
				var v []byte
				var pos []uint32
				for _, col := range row {
					v = append(v, col...)
					pos = append(pos, uint32(len(col)))
				}
				if err := insert.BindMultibytes(v, pos, len(pos), 3); err != nil {
					t.Fatal(err)
				}
			}

			got := readTextRows(t, conn, "SELECT a, b, c, typeof(a), typeof(b), typeof(c) FROM x ORDER BY rowid")
			want := make([][]string, 0, len(tt.want))
			for _, row := range tt.want {
				want = append(want, append(append([]string{}, row...), "text", "text", "text"))
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("rows mismatch\nwant: %q\ngot:  %q", want, got)
			}
		})
	}
}

func readTextRows(t *testing.T, conn *Conn, query string) [][]string {
	t.Helper()

	s, err := conn.Prepare(query)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var rows [][]string
	for {
		hasRow, err := s.Step()
		if err != nil {
			t.Fatal(err)
		}
		if !hasRow {
			return rows
		}

		row := make([]string, s.ColumnCount())
		for i := range row {
			row[i], _, err = s.ColumnText(i)
			if err != nil {
				t.Fatal(err)
			}
		}
		rows = append(rows, row)
	}
}
