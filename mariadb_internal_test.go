package testdbs

import "testing"

func TestNewMariadb(t *testing.T) {
	tcs := []struct {
		name  string
		image string
		want  string
	}{
		{name: "empty image uses the default", image: "", want: "mariadb:11.8"},
		{name: "custom image is used as given", image: "mariadb:10.11", want: "mariadb:10.11"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			db := NewMariadb(tc.image)
			if got := db.DbType(); got != DBTypeMariadb {
				t.Errorf("DbType() = %q, want %q", got, DBTypeMariadb)
			}
			my, ok := db.(*testDBMysql)
			if !ok {
				t.Fatalf("NewMariadb returned %T, want *testDBMysql", db)
			}
			if got := my.imageName(); got != tc.want {
				t.Errorf("image = %q, want %q", got, tc.want)
			}
		})
	}
}

// InitDBS runs MariaDB by default; MySQL is opt-in through NewMysql.
func TestDefaultDbsRunMariadbNotMysql(t *testing.T) {
	fast, slow := DefaultDbs()
	types := map[string]bool{}
	for _, db := range append(fast, slow...) {
		types[db.DbType()] = true
	}
	for _, want := range []string{DBTypeSqliteNOCgo, DBTypeSqliteCgo, DBTypeMariadb, DBTypePostgres} {
		if !types[want] {
			t.Errorf("default DBs miss %s", want)
		}
	}
	if types[DBTypeMysql] {
		t.Errorf("default DBs include %s, it must be opt-in", DBTypeMysql)
	}
}
