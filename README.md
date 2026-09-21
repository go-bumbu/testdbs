# testdbs
Utility package to start and stop different DBs in tests

# Usage

## Modify your tests

Modify the Main test function to initialize the DBs before the tests are run and to delete after

```
func TestMain(m *testing.M) {
	testdbs.InitDBS()
	// main block that runs tests
	code := m.Run()
	err := testdbs.Clean()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
```

And then in your tests you can iterate over the DBs

```
func TestMyFunction(t *testing.T) {
    for _, dbt := range testdbs.DBs() {
        t.Run(dbt.DbType(), func(t *testing.T) {
            db := dbt.Conn()
            // db is *gorm.DB
		})
	}
}
```

if you need to isolate DBs, e.g. run multiple tests on the same db/table
you can use  `ConnDbName("custom")` to crete a new database wit the passed name.

note: connections will be reused cross tests for every db name

```
func TestMyFunction(t *testing.T) {
    for _, dbt := range testdbs.DBs() {
        t.Run(dbt.DbType(), func(t *testing.T) {
            db := dbt.ConnDbName("custom")
            // db is *gorm.DB
		})
	}
}
```


## Custom DBs

`InitCustomDbs(fast, slow)` registers your own list instead of the default one:
the fast DBs always run, the slow ones only with `-alldbs` or `TESTDBS_ALL`.

### Postgres with another image

`NewPostgres(image)` runs any image that starts like the official `postgres`
image; an empty image keeps the default, `postgres:18`. For example, for the
pgvector extension:

```go
func TestMain(m *testing.M) {
	testdbs.InitCustomDbs(nil, []testdbs.TargetDb{
		testdbs.NewPostgres("pgvector/pgvector:pg18"),
	})
	code := m.Run()
	if err := testdbs.Clean(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
```

Extensions are per database: run `CREATE EXTENSION IF NOT EXISTS vector` on
every database you open, including each `ConnDbName` one.

### Only slow DBs registered

When every registered DB is a slow one, a plain `go test` selects none and
`DBs()` returns an empty list. Skip in that case:

```go
func TestSearch(t *testing.T) {
	dbs := testdbs.DBs()
	if len(dbs) == 0 {
		t.Skip("set TESTDBS_ALL=1 (needs Docker) to run the Postgres tests")
	}
	db := dbs[0].ConnDbName(t.Name()) // one database per test
	// ...
}
```

`DBs()` still panics if neither `InitDBS()` nor `InitCustomDbs()` ran.

## running tests

As a default calling `go test` will only start an embedded sqlite on a temp directory, to run the tests with
all the supported DBs you need to call `go test -alldbs`

Setting the env `TESTDBS_ALL` (any value) does the same as `-alldbs`. `-alldbs` is a flag of the test binary,
so `go test ./... -alldbs` fails in packages that don't import testdbs; in a module with such packages use
`TESTDBS_ALL=1 go test ./...` instead.

## sqlite

if you want to inspect the sqlite database after running the tests you can set the env `LOCAL_SQLITE` to true
and this will create the DB in the current working dir

```
export LOCAL_SQLITE=true
```