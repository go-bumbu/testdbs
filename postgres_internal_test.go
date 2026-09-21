package testdbs

import (
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
