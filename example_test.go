package testdbs_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-bumbu/testdbs"
)

// The examples mirror the README snippets, so the documented usage keeps
// compiling. The ones with an Output comment also run, against the DBs the
// package's TestMain selected; the TestMain ones only compile, since a test
// binary has a single TestMain.

// Start the default DBs once per package, from TestMain: SQLite always, and
// SQLite with CGo, MariaDB, and Postgres with -alldbs or TESTDBS_ALL.
func ExampleInitDBS() {
	testMain := func(m *testing.M) {
		testdbs.InitDBS()
		code := m.Run()
		// stop the containers and delete the SQLite files
		if err := testdbs.Clean(); err != nil {
			os.Exit(1)
		}
		os.Exit(code)
	}
	_ = testMain
}

// Add MySQL, which InitDBS leaves out, to the default list.
func ExampleDefaultDbs() {
	testMain := func(m *testing.M) {
		fast, slow := testdbs.DefaultDbs()
		testdbs.InitCustomDbs(fast, append(slow, testdbs.NewMysql("")))
		code := m.Run()
		if err := testdbs.Clean(); err != nil {
			os.Exit(1)
		}
		os.Exit(code)
	}
	_ = testMain
}

// Run Postgres from another image, here with pgvector, as the only DB. It is
// a slow DB, so a plain `go test` selects none and DBs() is empty.
func ExampleInitCustomDbs() {
	testMain := func(m *testing.M) {
		testdbs.InitCustomDbs(nil, []testdbs.TargetDb{
			testdbs.NewPostgres("pgvector/pgvector:pg18"),
		})
		code := m.Run()
		if err := testdbs.Clean(); err != nil {
			os.Exit(1)
		}
		os.Exit(code)
	}
	_ = testMain
}

type Note struct {
	ID   uint `gorm:"primaryKey"`
	Text string
}

// Run the same code once per DB. In a test, wrap the loop body in
// t.Run(dbt.DbType(), ...).
func ExampleDBs() {
	failed := 0
	for _, dbt := range testdbs.DBs() {
		// an isolated database, created on first use; dbt.Conn() is the
		// shared default one
		db := dbt.ConnDbName("ExampleDBs")
		if err := db.AutoMigrate(&Note{}); err != nil {
			failed++
			continue
		}
		db.Create(&Note{Text: "hello"})

		var count int64
		db.Model(&Note{}).Count(&count)
		if count != 1 {
			failed++
		}
	}
	fmt.Println("DBs that failed:", failed)
	// Output: DBs that failed: 0
}

// Open your own connection to the server DBs; SQLite has no connection
// parameters.
func ExampleConnParamsProvider() {
	incomplete := 0
	for _, dbt := range testdbs.DBs() {
		p, ok := dbt.(testdbs.ConnParamsProvider)
		if !ok {
			continue
		}
		params := p.ConnParamsDbName("ExampleConnParamsProvider")
		// build your own DSN from params here
		if params.Host == "" || params.Port == 0 || params.DbName == "" {
			incomplete++
		}
	}
	fmt.Println("incomplete params:", incomplete)
	// Output: incomplete params: 0
}
