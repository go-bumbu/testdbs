package testdbs

import (
	"errors"
	"flag"
	"fmt"
	sqliteNoCgo "github.com/glebarez/sqlite"
	"github.com/hashicorp/go-multierror"
	sqlitecgo "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"sync"
)

// return the path for a db name
func dbPath(name, suffix, tmpdir string, isLocal bool) string {
	if name == "" {
		name = defaultDbName
	}
	dbName := fmt.Sprintf("%s.%s.sqlite", name, suffix)
	if isLocal {
		tmpdir = "./" + testDbDir + "_sqlite"
		if err := os.MkdirAll(tmpdir, 0750); err != nil {
			panic(fmt.Sprintf("error creating local temporary directory: %v", err))
		}
	}
	return fmt.Sprintf("%s/%s", tmpdir, dbName)
}

const testDbDir = "testdbs"

func mkTmpDir() string {
	dir, err := os.MkdirTemp("", testDbDir+"_sqlite")
	if err != nil {
		panic(fmt.Sprintf("error creating temporary directory: %v", err))
	}
	return dir
}

// removeTmpDir deletes a temp dir made by mkTmpDir, refusing any other path.
func removeTmpDir(dir string) error {
	if !strings.Contains(dir, testDbDir) {
		panic("refusing to delete the dir since it does not seem to be from testdbs")
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("error cleaning up temporary directory: %w", err)
	}
	return nil
}

// openSqlite opens a fresh DB at file, deleting what a previous run left there.
func openSqlite(file string, open func(dsn string) gorm.Dialector, l logger.Interface) *gorm.DB {
	if _, err := os.Stat(file); err == nil {
		err = os.RemoveAll(file)
		if err != nil {
			panic(fmt.Sprintf("error while removing dbfile: %v", err))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		panic(fmt.Sprintf(" while doing stat on dbfile: %v", err))
	}

	db, err := gorm.Open(open(file), &gorm.Config{
		Logger: l,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to open test database: %v", err))
	}
	return db
}

// closeSqlite closes the named DB and drops it from pool. Outside local mode
// it also deletes the DB file; in local mode the file stays for inspection.
func closeSqlite(pool map[string]*gorm.DB, name, file string, isLocal bool) error {
	dbConn, exists := pool[name]
	if !exists {
		return fmt.Errorf("db connection with name %s not found", name)
	}
	db, err := dbConn.DB()
	if err != nil {
		return err
	}
	if err = db.Close(); err != nil {
		return err
	}
	delete(pool, name)

	if !isLocal {
		if err = os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error removing db file: %w", err)
		}
	}
	return nil
}

// Add a flag to run sqlite on the local dir instead of on the tmpdir
var sqliteLoca *bool

func init() {
	sqliteLoca = flag.Bool("localsqlite", false, "create sqlite DBs in the CWD")
}

func sqliteLocal() bool {
	if sqliteLoca == nil {
		panic("testing: sqliteLocal called before Init")
	}
	// Catch code that calls this from TestMain without first calling flag.Parse.
	if !flag.Parsed() {
		panic("testing: sqliteLocal called before Parse")
	}
	return *sqliteLoca
}

// ===============================================================================
// Sqlite without CGO
// ===============================================================================

const (
	noCgoSqliteSuffix = "no_CGO"
	DBTypeSqliteNOCgo = "SqliteNoCgo"
)

// SqliteNoCgo is safe for concurrent use: mu guards pool, and ConnDbName holds it
// while creating a DB so parallel callers asking for one name share one pool.
type SqliteNoCgo struct {
	mu      sync.Mutex
	dir     string
	isLocal bool
	logger  logger.Interface
	pool    map[string]*gorm.DB
}

func (c *SqliteNoCgo) DbType() string {
	return DBTypeSqliteNOCgo
}
func (c *SqliteNoCgo) Init(logger logger.Interface) {
	c.logger = logger
	c.pool = map[string]*gorm.DB{}

	_, localSqliteEnv := os.LookupEnv(LocalSqliteEnv)
	if localSqliteEnv || sqliteLocal() {
		c.isLocal = true
		c.dir = "./"
	} else {
		c.dir = mkTmpDir()
		c.isLocal = false
	}
}
func (c *SqliteNoCgo) Conn() *gorm.DB {
	return c.ConnDbName(defaultDbName)
}

func (c *SqliteNoCgo) ConnDbName(name string) *gorm.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	name = normalizeDbName(name)
	dbConn, exists := c.pool[name]
	if exists {
		return dbConn
	}
	db := openSqlite(dbPath(name, noCgoSqliteSuffix, c.dir, c.isLocal), sqliteNoCgo.Open, c.logger)
	c.pool[name] = db
	return db
}

func (c *SqliteNoCgo) Close(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return closeSqlite(c.pool, name, dbPath(name, noCgoSqliteSuffix, c.dir, c.isLocal), c.isLocal)
}

// CloseAll closes every DB and, outside local mode, deletes the temp dir.
func (c *SqliteNoCgo) CloseAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var merr error
	for name := range c.pool {
		err := closeSqlite(c.pool, name, dbPath(name, noCgoSqliteSuffix, c.dir, c.isLocal), c.isLocal)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	if !c.isLocal {
		if err := removeTmpDir(c.dir); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	return merr
}

// ===============================================================================
// Sqlite using CGO
// ===============================================================================

const (
	CgoSqliteSuffix = "with_CGO"
	DBTypeSqliteCgo = "SqliteWithCgo"
)

// SqliteCgo is safe for concurrent use: mu guards pool, and ConnDbName holds it
// while creating a DB so parallel callers asking for one name share one pool.
type SqliteCgo struct {
	mu      sync.Mutex
	dir     string
	isLocal bool
	logger  logger.Interface
	pool    map[string]*gorm.DB
}

func (c *SqliteCgo) DbType() string {
	return DBTypeSqliteCgo
}
func (c *SqliteCgo) Init(logger logger.Interface) {
	c.logger = logger
	c.pool = map[string]*gorm.DB{}

	_, localSqliteEnv := os.LookupEnv(LocalSqliteEnv)
	if localSqliteEnv || sqliteLocal() {
		c.isLocal = true
		c.dir = "./"
	} else {
		c.dir = mkTmpDir()
		c.isLocal = false
	}
}
func (c *SqliteCgo) Conn() *gorm.DB {
	return c.ConnDbName(defaultDbName)
}

func (c *SqliteCgo) ConnDbName(name string) *gorm.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	name = normalizeDbName(name)
	dbConn, exists := c.pool[name]
	if exists {
		return dbConn
	}
	db := openSqlite(dbPath(name, CgoSqliteSuffix, c.dir, c.isLocal), sqlitecgo.Open, c.logger)
	c.pool[name] = db
	return db
}

func (c *SqliteCgo) Close(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return closeSqlite(c.pool, name, dbPath(name, CgoSqliteSuffix, c.dir, c.isLocal), c.isLocal)
}

// CloseAll closes every DB and, outside local mode, deletes the temp dir.
func (c *SqliteCgo) CloseAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var merr error
	for name := range c.pool {
		err := closeSqlite(c.pool, name, dbPath(name, CgoSqliteSuffix, c.dir, c.isLocal), c.isLocal)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	if !c.isLocal {
		if err := removeTmpDir(c.dir); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	return merr
}
