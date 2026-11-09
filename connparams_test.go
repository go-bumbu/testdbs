package testdbs_test

import (
	"fmt"
	"testing"

	"github.com/go-bumbu/testdbs"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openFromParams opens a new connection from the exported parameters only,
// the way a caller with its own DSN builder or gorm.Config would.
func openFromParams(t *testing.T, dbType string, p testdbs.ConnParams) *gorm.DB {
	t.Helper()
	var dialector gorm.Dialector
	switch dbType {
	case testdbs.DBTypePostgres:
		dialector = postgres.Open(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			p.Host, p.Port, p.User, p.Password, p.DbName, p.SslMode))
	case testdbs.DBTypeMysql:
		dialector = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=True",
			p.User, p.Password, p.Host, p.Port, p.DbName))
	default:
		t.Fatalf("no dialector for %s", dbType)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Discard, TranslateError: true})
	if err != nil {
		t.Fatalf("open from params: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Errorf("get sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	return db
}

func TestConnParams(t *testing.T) {
	ran := false
	for _, dbt := range testdbs.DBs() {
		p, ok := dbt.(testdbs.ConnParamsProvider)
		switch dbt.DbType() {
		case testdbs.DBTypePostgres, testdbs.DBTypeMysql:
			if !ok {
				t.Fatalf("%s target does not implement ConnParamsProvider", dbt.DbType())
			}
		default:
			continue
		}
		ran = true

		t.Run(dbt.DbType(), func(t *testing.T) {
			t.Run("default db", func(t *testing.T) {
				params := p.ConnParams()
				if params.Host == "" || params.Port == 0 || params.User == "" || params.Password == "" || params.DbName == "" {
					t.Fatalf("incomplete params: %+v", params)
				}
				if dbt.DbType() == testdbs.DBTypePostgres && params.SslMode != "disable" {
					t.Errorf("SslMode = %q, want disable", params.SslMode)
				}

				var one int
				if err := openFromParams(t, dbt.DbType(), params).Raw("SELECT 1").Scan(&one).Error; err != nil {
					t.Fatalf("query: %v", err)
				}
				if one != 1 {
					t.Errorf("SELECT 1 = %d, want 1", one)
				}
			})

			// A row written through ConnDbName must be visible on a connection
			// opened from ConnParamsDbName with the same name: same database.
			t.Run("named db", func(t *testing.T) {
				name := t.Name()
				if err := dbt.ConnDbName(name).AutoMigrate(&Item{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
				written := Item{Name: "via ConnDbName"}
				if err := dbt.ConnDbName(name).Create(&written).Error; err != nil {
					t.Fatalf("create: %v", err)
				}

				params := p.ConnParamsDbName(name)
				if params.DbName == p.ConnParams().DbName {
					t.Fatalf("named db shares the default db name %q", params.DbName)
				}
				var read Item
				if err := openFromParams(t, dbt.DbType(), params).First(&read, written.ID).Error; err != nil {
					t.Fatalf("read: %v", err)
				}
				if read != written {
					t.Errorf("read %+v, want %+v", read, written)
				}
			})

			// ConnParamsDbName alone must create the database.
			t.Run("fresh named db", func(t *testing.T) {
				db := openFromParams(t, dbt.DbType(), p.ConnParamsDbName(t.Name()))
				if err := db.AutoMigrate(&Item{}); err != nil {
					t.Fatalf("automigrate: %v", err)
				}
			})
		})
	}
	if !ran {
		t.Skip("no Postgres or MySQL target: run with -alldbs or TESTDBS_ALL")
	}
}

func TestConnParamsNotOnSqlite(t *testing.T) {
	for _, dbt := range testdbs.DBs() {
		if dbt.DbType() == testdbs.DBTypePostgres || dbt.DbType() == testdbs.DBTypeMysql {
			continue
		}
		if _, ok := dbt.(testdbs.ConnParamsProvider); ok {
			t.Errorf("%s implements ConnParamsProvider, it has no server to connect to", dbt.DbType())
		}
	}
}
