package store_test

import (
	"flag"
	"os"
	"slices"
	"testing"

	"example.com/consumer/store"
	"github.com/go-bumbu/testdbs"
)

// TestMain as the README shows it.
func TestMain(m *testing.M) {
	testdbs.InitDBS()
	code := m.Run()
	if err := testdbs.Clean(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}

// A plain `go test` runs only the embedded SQLite; TESTDBS_ALL or -alldbs
// adds the Docker DBs.
func TestSelectedDbs(t *testing.T) {
	want := []string{testdbs.DBTypeSqliteNOCgo}
	_, env := os.LookupEnv(testdbs.RunAllDBsEnv)
	if env || flag.Lookup("alldbs").Value.String() == "true" {
		want = append(want, testdbs.DBTypeSqliteCgo, testdbs.DBTypeMariadb, testdbs.DBTypePostgres)
	}
	var got []string
	for _, dbt := range testdbs.DBs() {
		got = append(got, dbt.DbType())
	}
	if !slices.Equal(got, want) {
		t.Errorf("DBs() = %v, want %v", got, want)
	}
}

func TestAddGet(t *testing.T) {
	for _, dbt := range testdbs.DBs() {
		t.Run(dbt.DbType(), func(t *testing.T) {
			db := dbt.ConnDbName(t.Name())
			if err := store.Migrate(db); err != nil {
				t.Fatalf("migrate: %v", err)
			}
			id, err := store.Add(db, "hello")
			if err != nil {
				t.Fatalf("add: %v", err)
			}
			n, err := store.Get(db, id)
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if n.Text != "hello" {
				t.Errorf("Text = %q, want hello", n.Text)
			}
		})
	}
}
