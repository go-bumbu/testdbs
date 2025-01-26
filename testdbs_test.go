package testdbs_test

import (
	"github.com/go-bumbu/testdbs"
	"github.com/google/go-cmp/cmp"
	"log"
	"os"
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
			})
		}
	})

}
