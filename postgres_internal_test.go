package testdbs

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func TestNewPostgresImage(t *testing.T) {
	tcs := []struct {
		name  string
		image string
		want  string
	}{
		{name: "empty image uses the default", image: "", want: "postgres:13"},
		{name: "custom image is used as given", image: "pgvector/pgvector:pg17", want: "pgvector/pgvector:pg17"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			db := NewPostgres(tc.image)
			pg, ok := db.(*testDBPostgres)
			if !ok {
				t.Fatalf("NewPostgres returned %T, want *testDBPostgres", db)
			}
			if got := pg.imageName(); got != tc.want {
				t.Errorf("image = %q, want %q", got, tc.want)
			}
		})
	}

	// InitDBS builds the zero value directly; it must keep the default.
	t.Run("zero value uses the default", func(t *testing.T) {
		if got := (&testDBPostgres{}).imageName(); got != "postgres:13" {
			t.Errorf("image = %q, want %q", got, "postgres:13")
		}
	})
}

// TestNewPostgresPgvector proves the use case behind NewPostgres: the vector
// extension works on the default database and on a fresh ConnDbName one.
// Extensions are per database, so both need checking.
func TestNewPostgresPgvector(t *testing.T) {
	if !slowDBsEnabled() {
		t.Skip("starts a container: run with -alldbs or TESTDBS_ALL")
	}
	pg := NewPostgres("pgvector/pgvector:pg17")
	pg.Init(logger.Discard)
	t.Cleanup(func() {
		if err := pg.CloseAll(); err != nil {
			t.Errorf("CloseAll: %v", err)
		}
	})

	tcs := []struct {
		name string
		db   *gorm.DB
	}{
		{name: "Conn", db: pg.Conn()},
		{name: "ConnDbName", db: pg.ConnDbName(t.Name())},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
				t.Fatalf("create extension: %v", err)
			}
			var dims int
			if err := tc.db.Raw("SELECT vector_dims('[1,2,3]'::vector)").Scan(&dims).Error; err != nil {
				t.Fatalf("query vector: %v", err)
			}
			if dims != 3 {
				t.Errorf("vector_dims = %d, want 3", dims)
			}
		})
	}
}
