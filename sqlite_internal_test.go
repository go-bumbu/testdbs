package testdbs

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
)

// sqliteTargets returns constructors for fresh SQLite targets, keyed by type.
// The cgo one is a slow DB in InitDBS, so it only runs with -alldbs or
// TESTDBS_ALL.
func sqliteTargets() map[string]func() TargetDb {
	targets := map[string]func() TargetDb{
		DBTypeSqliteNOCgo: func() TargetDb { return &SqliteNoCgo{} },
	}
	if slowDBsEnabled() {
		targets[DBTypeSqliteCgo] = func() TargetDb { return &SqliteCgo{} }
	}
	return targets
}

// sqliteFile returns the path of the file behind a SQLite connection.
func sqliteFile(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var file string
	if err := db.Raw("SELECT file FROM pragma_database_list WHERE name = 'main'").Scan(&file).Error; err != nil {
		t.Fatalf("look up db file: %v", err)
	}
	if file == "" {
		t.Fatal("no file for the main database")
	}
	return file
}

func TestSqliteCloseAllRemovesTempFiles(t *testing.T) {
	for name, newDb := range sqliteTargets() {
		t.Run(name, func(t *testing.T) {
			db := newDb()
			db.Init(logger.Discard)
			file := sqliteFile(t, db.ConnDbName("closeme"))

			if err := db.Close("missing"); err == nil {
				t.Error("Close(missing) = nil, want an error")
			}
			if err := db.CloseAll(); err != nil {
				t.Fatalf("CloseAll() = %v, want nil", err)
			}
			if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("%s still exists after CloseAll (stat error: %v)", file, err)
			}
		})
	}
}

// With LOCAL_SQLITE the files stay in the working directory for inspection,
// and the next run with the same name starts from an empty database.
func TestSqliteLocalKeepsFilesUntilNextRun(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv(LocalSqliteEnv, "true")

	for name, newDb := range sqliteTargets() {
		t.Run(name, func(t *testing.T) {
			first := newDb()
			first.Init(logger.Discard)
			conn := first.ConnDbName("keep")
			if err := conn.Exec("CREATE TABLE marker (id INTEGER)").Error; err != nil {
				t.Fatalf("create table: %v", err)
			}
			file := sqliteFile(t, conn)
			if err := first.CloseAll(); err != nil {
				t.Fatalf("CloseAll() = %v, want nil", err)
			}
			if _, err := os.Stat(file); err != nil {
				t.Fatalf("local db file should survive CloseAll: %v", err)
			}

			second := newDb()
			second.Init(logger.Discard)
			t.Cleanup(func() {
				if err := second.CloseAll(); err != nil {
					t.Errorf("CloseAll() = %v, want nil", err)
				}
			})
			if second.ConnDbName("keep").Migrator().HasTable("marker") {
				t.Error("the next run still sees the previous run's table")
			}
		})
	}
}
