# testdbs

[![Go Reference](https://pkg.go.dev/badge/github.com/go-bumbu/testdbs.svg)](https://pkg.go.dev/github.com/go-bumbu/testdbs)
[![CI](https://github.com/go-bumbu/testdbs/actions/workflows/test.yml/badge.svg)](https://github.com/go-bumbu/testdbs/actions/workflows/test.yml)
[![License: LGPL-3.0](https://img.shields.io/badge/license-LGPL--3.0-blue.svg)](LICENSE)

testdbs starts SQLite, MariaDB, and Postgres (and MySQL, if you ask for it) for your Go tests and runs each test against every one of them, through GORM.
A plain `go test` uses an embedded SQLite; add `-alldbs` to run the full set in Docker containers.

## Features

- Run the same test against SQLite, MariaDB 11.8, and Postgres 18, plus MySQL 9.7 when you opt in
- Keep `go test` fast and Docker-free; turn on the full set with one flag
- Give each test its own database, so tests don't share tables
- Run the database containers in memory with durability off (no fsync), so the suite runs faster
- Open your own connection to MariaDB, MySQL, or Postgres, with your own DSN or `gorm.Config`
- Test against another Postgres, MariaDB, or MySQL image, such as pgvector or MySQL 8.4

## Add to your project

```sh
go get github.com/go-bumbu/testdbs@latest
```

The MariaDB, MySQL, Postgres, and CGo SQLite targets need Docker, and only run with `-alldbs`.

### Start the DBs once per package

```go
func TestMain(m *testing.M) {
	// start the DBs before the tests run
	testdbs.InitDBS()
	code := m.Run()
	// stop the containers and delete the SQLite files
	if err := testdbs.Clean(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
```

### Run a test on every DB

```go
func TestMyFunction(t *testing.T) {
	// run the same test once per DB
	for _, dbt := range testdbs.DBs() {
		t.Run(dbt.DbType(), func(t *testing.T) {
			// a *gorm.DB connected to the default database
			db := dbt.Conn()
		})
	}
}
```

To isolate a test, use `dbt.ConnDbName(t.Name())` instead of `dbt.Conn()`: it
creates a database with that name on first use and reuses its connection after.

### Open your own connection

To test code that builds its own DSN or `gorm.Config`, Postgres, MariaDB, and
MySQL expose their connection parameters. SQLite does not.

```go
func TestOpen(t *testing.T) {
	for _, dbt := range testdbs.DBs() {
		// only the server DBs expose connection parameters
		p, ok := dbt.(testdbs.ConnParamsProvider)
		if !ok {
			continue
		}
		t.Run(dbt.DbType(), func(t *testing.T) {
			// Host, Port, User, Password, DbName, and SslMode of an isolated database
			params := p.ConnParamsDbName(t.Name()) // or p.ConnParams() for the default one
			// open your own connection here, and close it when done
		})
	}
}
```

`ConnParamsDbName(name)` points at the same database as `ConnDbName(name)`.
`SslMode` is `disable` for Postgres and empty for MariaDB and MySQL.

### Also run MySQL

`InitDBS()` runs MariaDB, not MySQL. MariaDB shares MySQL's driver and protocol,
so `dbt.DbType()` (`testdbs.DBTypeMariadb` or `testdbs.DBTypeMysql`) is how a
test tells them apart. To run MySQL as well, add it to the default list:

```go
func TestMain(m *testing.M) {
	fast, slow := testdbs.DefaultDbs()
	testdbs.InitCustomDbs(fast, append(slow, testdbs.NewMysql("")))
	code := m.Run()
	if err := testdbs.Clean(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
```

### Use another Postgres, MariaDB, or MySQL image

`InitCustomDbs(fast, slow)` replaces the default list: fast DBs always run, slow
ones only with `-alldbs`. `NewPostgres(image)` runs any image that starts like
the official `postgres` one, `NewMariadb(image)` like the official `mariadb`
one (e.g. `"mariadb:10.11"`), and `NewMysql(image)` like the official `mysql`
one, 8.0.30 or later (e.g. `"mysql:8.4"`). An empty image keeps the default.

```go
func TestMain(m *testing.M) {
	// Postgres with pgvector, as the only (slow) DB
	testdbs.InitCustomDbs(nil, []testdbs.TargetDb{
		testdbs.NewPostgres("pgvector/pgvector:pg18"),
	})
	code := m.Run()
	if err := testdbs.Clean(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}

func TestSearch(t *testing.T) {
	// a plain `go test` selects no slow DBs, so skip
	dbs := testdbs.DBs()
	if len(dbs) == 0 {
		t.Skip("set TESTDBS_ALL=1 (needs Docker) to run the Postgres tests")
	}
	// extensions are per database: create them in every database you open
	db := dbs[0].ConnDbName(t.Name())
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
}
```

### Run your tests

```sh
go test ./...                 # embedded SQLite only
go test ./... -alldbs         # every DB, needs Docker
TESTDBS_ALL=1 go test ./...   # same, for modules where some packages don't import testdbs
LOCAL_SQLITE=1 go test ./...  # keep the SQLite files in the working dir to inspect them
```

Full API: [pkg.go.dev/github.com/go-bumbu/testdbs](https://pkg.go.dev/github.com/go-bumbu/testdbs).

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md).

Licensed under the [GNU LGPL v3](LICENSE).
