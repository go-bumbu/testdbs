package testdbs

import (
	"flag"
	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

type TargetDb interface {
	DbType() string
	Init(logger logger.Interface)
	Conn() *gorm.DB
	ConnDbName(name string) *gorm.DB
	Close(name string) error
	CloseAll() error
}

const (
	LocalSqliteEnv = "LOCAL_SQLITE"
	RunAllDBsEnv   = "TESTDBS_ALL"
	defaultDbName  = "testdbDefault"
)

// InitDBS initializes the default DBs: SQLite without CGo always, and SQLite
// with CGo, MariaDB, and Postgres with -alldbs or TESTDBS_ALL. MySQL is not in
// the default set; pass NewMysql to InitCustomDbs to run it. Call it once,
// from TestMain.
func InitDBS() {
	InitCustomDbs(DefaultDbs())
}

// DefaultDbs returns new instances of the fast and slow DBs InitDBS runs, to
// extend them before calling InitCustomDbs, e.g. to also run MySQL:
//
//	fast, slow := testdbs.DefaultDbs()
//	testdbs.InitCustomDbs(fast, append(slow, testdbs.NewMysql("")))
func DefaultDbs() (fast, slow []TargetDb) {
	fast = []TargetDb{&SqliteNoCgo{}}
	slow = []TargetDb{
		&SqliteCgo{},
		NewMariadb(""),
		NewPostgres(""),
	}
	return fast, slow
}

// InitCustomDbs initializes the given DBs and makes them available through
// DBs(). fastDbs always run; longDBs run only when the -alldbs flag or the
// TESTDBS_ALL env var is set. Call it once, from TestMain.
func InitCustomDbs(fastDbs, longDBs []TargetDb) {

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	flag.Parse()

	dbs := fastDbs

	// also run slow DBs
	if slowDBsEnabled() {
		for _, db := range longDBs {
			if !slices.Contains(dbs, db) {
				dbs = append(dbs, db)
			}
		}
	}

	for _, db := range dbs {
		db.Init(gormLogger)
		targetDBS = append(targetDBS, db)
	}
	initialized = true
}

var (
	targetDBS = []TargetDb{}
	// initialized records that InitDBS or InitCustomDbs ran, even when it
	// selected no DBs.
	initialized bool
)

// DBs returns the DBs selected by InitDBS or InitCustomDbs. The list is empty
// when only slow DBs were registered and neither -alldbs nor TESTDBS_ALL is
// set, so tests should skip in that case. DBs panics if no init function ran.
func DBs() []TargetDb {
	if !initialized {
		panic("testdbs were not initialized, run InitDBS() or InitCustomDbs() before calling DBs()")
	}
	return targetDBS
}

// Clean closes all db connections and deletes related test files
func Clean() error {
	var merr error
	for _, db := range targetDBS {
		err := db.CloseAll()
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	return merr
}

// Flag to run fast DBs or all DBs
var runAllDbs *bool

func init() {
	runAllDbs = flag.Bool("alldbs", false, "run the tests on all available DBs")
}
func testAll() bool {
	if runAllDbs == nil {
		panic("testing: testAll called before Init")
	}
	// Catch code that calls this from TestMain without first calling flag.Parse.
	if !flag.Parsed() {
		panic("testing: testAll called before Parse")
	}
	return *runAllDbs
}

// slowDBsEnabled reports whether the slow DBs were requested, with the
// -alldbs flag or the TESTDBS_ALL env var.
func slowDBsEnabled() bool {
	_, env := os.LookupEnv(RunAllDBsEnv)
	return env || testAll()
}

func normalizeDbName(input string) string {
	// Limit length to 64 characters
	if len(input) > 64 {
		input = input[:64]
	}
	input = strings.Trim(input, "/")
	input = strings.Trim(input, "\\")
	input = strings.Trim(input, ".")

	// Replace '/', '\', and '.' with '-'
	replacer := strings.NewReplacer("/", "", "\\", "", ".", "")
	input = replacer.Replace(input)

	// Replace spaces with '-'
	input = strings.ReplaceAll(input, " ", "")

	// remove
	input = strings.TrimPrefix(input, "-")
	input = strings.TrimSuffix(input, "-")

	// Remove any characters that are not permitted in file names
	illegalChars := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	input = illegalChars.ReplaceAllString(input, "")

	input = strings.ToLower(input)

	return input
}
