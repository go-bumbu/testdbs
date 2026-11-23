package testdbs

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewMysqlImage(t *testing.T) {
	tcs := []struct {
		name  string
		image string
		want  string
	}{
		{name: "empty image uses the default", image: "", want: "mysql:9.7"},
		{name: "custom image is used as given", image: "mysql:8.4", want: "mysql:8.4"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			db := NewMysql(tc.image)
			my, ok := db.(*testDBMysql)
			if !ok {
				t.Fatalf("NewMysql returned %T, want *testDBMysql", db)
			}
			if got := my.imageName(); got != tc.want {
				t.Errorf("image = %q, want %q", got, tc.want)
			}
		})
	}

	// InitDBS builds the zero value directly; it must keep the default.
	t.Run("zero value uses the default", func(t *testing.T) {
		if got := (&testDBMysql{}).imageName(); got != "mysql:9.7" {
			t.Errorf("image = %q, want %q", got, "mysql:9.7")
		}
	})
}

// TestNewMysqlOtherImage proves NewMysql runs another image end to end: the
// container starts with the durability flags, and both the default and a
// fresh ConnDbName database work on it.
func TestNewMysqlOtherImage(t *testing.T) {
	if !slowDBsEnabled() {
		t.Skip("starts a container: run with -alldbs or TESTDBS_ALL")
	}
	const image = "mysql:8.4"
	my := NewMysql(image)
	my.Init(logger.Discard)
	t.Cleanup(func() {
		if err := my.CloseAll(); err != nil {
			t.Errorf("CloseAll: %v", err)
		}
	})

	var version string
	if err := my.Conn().Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		t.Fatalf("query version: %v", err)
	}
	if len(version) < 4 || version[:4] != "8.4." {
		t.Errorf("VERSION() = %q, want 8.4.x from %s", version, image)
	}

	tcs := []struct {
		name string
		db   *gorm.DB
	}{
		{name: "Conn", db: my.Conn()},
		{name: "ConnDbName", db: my.ConnDbName(t.Name())},
	}
	type row struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.db.AutoMigrate(&row{}); err != nil {
				t.Fatalf("automigrate: %v", err)
			}
			if err := tc.db.Create(&row{Name: "x"}).Error; err != nil {
				t.Fatalf("create: %v", err)
			}
		})
	}
}
