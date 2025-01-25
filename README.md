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
    testdbs.Clean()
    os.Exit(code)
}
```

And then in your tests you can iterate over the DBs

```
func TestScan(t *testing.T) {
    for _, db := range testdbs.TargetDBS {
        t.Run(db.Name, func(t *testing.T) {
            // do something with db, db is *Gorm.DB
		})
	}
}
```

## Calling

As a default calling `go test` will only start an embedded sqlite on a temp directory, to run the tests with
all the supported DBs you need to call `go test -alldbs`