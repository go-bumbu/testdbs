package testdbs

import (
	"flag"
	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"slices"
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

func InitDBS() {
	fast := []TargetDb{&SqliteNoCgo{}}
	long := []TargetDb{
		&SqliteCgo{},
		&testDBMysql{},
		&testDBPostgres{},
	}
	InitCustomDbs(fast, long)
}

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

	allDbs := false
	_, testAllEnv := os.LookupEnv(RunAllDBsEnv)
	if testAllEnv || testAll() {
		allDbs = true
	}

	dbs := fastDbs

	// also run slow DBs
	if allDbs {
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

}

var targetDBS = []TargetDb{}

func DBs() []TargetDb {
	if len(targetDBS) == 0 {
		panic("testdbs were not initialized, run InitDBS() before calling DBs()")
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
