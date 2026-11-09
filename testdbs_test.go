package testdbs_test

import (
	"fmt"
	"github.com/go-bumbu/testdbs"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	testdbs.InitDBS()
	// main block that runs tests
	code := m.Run()
	err := testdbs.Clean()
	if err != nil {
		os.Exit(1)
	}

	//check for routine leaks
	opts := []goleak.Option{
		// Not ideal: ignore the SQL connection opener.
		// TODO: explore the impact of the connectionOpener routine being still still running after closing the db connection
		goleak.IgnoreAnyFunction("database/sql.(*DB).connectionOpener"),
	}
	if err := goleak.Find(opts...); err != nil {
		fmt.Printf("found routine leak: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

type Item struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func TestStartDbNew(t *testing.T) {
	t.Run("defaultDB", func(t *testing.T) {
		for _, dbt := range testdbs.DBs() {
			t.Run(dbt.DbType(), func(t *testing.T) {
				db := dbt.Conn()

				err := db.AutoMigrate(&Item{})
				if err != nil {
					t.Fatalf("error in automigrate: %s", err)
				}

				writtenItem := Item{Name: "Sample Item"}
				result := db.Create(&writtenItem) // Write item to the database
				if result.Error != nil {
					log.Fatalf("Failed to create item: %v", result.Error)
				}

				var readItem Item
				db.First(&readItem, writtenItem.ID) // Fetch the first item by ID

				if diff := cmp.Diff(writtenItem, readItem, cmp.AllowUnexported(Item{})); diff != "" {
					t.Errorf("Mismatch (-written +read):\n%s", diff)
				}
			})
		}
	})

	t.Run("customDb", func(t *testing.T) {
		for _, dbt := range testdbs.DBs() {
			t.Run(dbt.DbType(), func(t *testing.T) {
				db := dbt.ConnDbName("custom")

				err := db.AutoMigrate(&Item{})
				if err != nil {
					t.Fatalf("error in automigrate: %s", err)
				}

				writtenItem := Item{Name: "Sample Item"}
				result := db.Create(&writtenItem) // Write item to the database
				if result.Error != nil {
					log.Fatalf("Failed to create item: %v", result.Error)
				}

				var readItem Item
				db.First(&readItem, writtenItem.ID) // Fetch the first item by ID

				if diff := cmp.Diff(writtenItem, readItem, cmp.AllowUnexported(Item{})); diff != "" {
					t.Errorf("Mismatch (-written +read):\n%s", diff)
				}

				// second DB conn
				db2 := dbt.ConnDbName("custom")

				var readItem2 Item
				db2.First(&readItem2, writtenItem.ID) // Fetch the first item by ID

				if diff := cmp.Diff(writtenItem, readItem2, cmp.AllowUnexported(Item{})); diff != "" {
					t.Errorf("Mismatch (-written +read):\n%s", diff)
				}

				// 3rd DB conn
				db3 := dbt.ConnDbName("custom2")
				err = db3.AutoMigrate(&Item{})
				if err != nil {
					t.Fatalf("error in automigrate: %s", err)
				}

				var readItem3 Item
				db3.First(&readItem3, writtenItem.ID) // Fetch the first item by ID

				if diff := cmp.Diff(Item{}, readItem3, cmp.AllowUnexported(Item{})); diff != "" {
					t.Errorf("Mismatch (-written +read):\n%s", diff)
				}
			})
		}
	})

	t.Run("customDbTestName", func(t *testing.T) {
		for _, dbt := range testdbs.DBs() {
			t.Run(dbt.DbType(), func(t *testing.T) {
				db := dbt.ConnDbName(t.Name())

				err := db.AutoMigrate(&Item{})
				if err != nil {
					t.Fatalf("error in automigrate: %s", err)
				}

				writtenItem := Item{Name: "Sample Item"}
				result := db.Create(&writtenItem) // Write item to the database
				if result.Error != nil {
					log.Fatalf("Failed to create item: %v", result.Error)
				}

				var readItem Item
				db.First(&readItem, writtenItem.ID) // Fetch the first item by ID

				if diff := cmp.Diff(writtenItem, readItem, cmp.AllowUnexported(Item{})); diff != "" {
					t.Errorf("Mismatch (-written +read):\n%s", diff)
				}

			})
		}
	})
}

func TestCloseNamedDb(t *testing.T) {
	for _, dbt := range testdbs.DBs() {
		t.Run(dbt.DbType(), func(t *testing.T) {
			dbt.ConnDbName("closeme")
			if err := dbt.Close("closeme"); err != nil {
				t.Errorf("Close(closeme) = %v, want nil", err)
			}
			if err := dbt.Close("missing"); err == nil {
				t.Error("Close(missing) = nil, want an error")
			}
		})
	}
}

// Parallel subtests must be able to share a target: each asks for its own DB
// and for one DB they all share, and the shared name yields a single pool.
func TestConnDbNameConcurrent(t *testing.T) {
	const workers = 8
	for _, dbt := range testdbs.DBs() {
		t.Run(dbt.DbType(), func(t *testing.T) {
			var mu sync.Mutex
			shared := map[*gorm.DB]bool{}

			t.Run("workers", func(t *testing.T) {
				for i := range workers {
					t.Run(fmt.Sprintf("w%d", i), func(t *testing.T) {
						t.Parallel()
						own := dbt.ConnDbName(t.Name())
						if err := own.AutoMigrate(&Item{}); err != nil {
							t.Fatalf("automigrate own db: %v", err)
						}
						db := dbt.ConnDbName("concurrentshared")
						if err := db.Exec("SELECT 1").Error; err != nil {
							t.Fatalf("query shared db: %v", err)
						}
						mu.Lock()
						shared[db] = true
						mu.Unlock()
					})
				}
			})

			if len(shared) != 1 {
				t.Errorf("ConnDbName(concurrentshared) returned %d distinct pools, want 1", len(shared))
			}
		})
	}
}
