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
