package testdbs_test

import (
	"testing"

	"github.com/go-bumbu/testdbs"
)

// The containers run without durability for speed; check the settings reached
// the server, a typo in a flag would otherwise go unnoticed.
func TestDurabilityOff(t *testing.T) {
	tcs := map[string][]struct{ query, want string }{
		testdbs.DBTypePostgres: {
			{"SHOW fsync", "off"},
			{"SHOW synchronous_commit", "off"},
			{"SHOW full_page_writes", "off"},
			{"SHOW data_directory", "/pgdata"},
		},
		testdbs.DBTypeMysql: {
			{"SELECT @@innodb_flush_log_at_trx_commit", "0"},
			{"SELECT @@innodb_doublewrite", "OFF"},
			{"SELECT @@log_bin", "0"},
		},
		testdbs.DBTypeMariadb: {
			{"SELECT @@innodb_flush_log_at_trx_commit", "0"},
			{"SELECT @@innodb_doublewrite", "OFF"},
			{"SELECT @@log_bin", "0"},
		},
	}
	ran := false
	for _, dbt := range testdbs.DBs() {
		checks, ok := tcs[dbt.DbType()]
		if !ok {
			continue
		}
		ran = true
		t.Run(dbt.DbType(), func(t *testing.T) {
			for _, c := range checks {
				var got string
				if err := dbt.Conn().Raw(c.query).Scan(&got).Error; err != nil {
					t.Fatalf("%s: %v", c.query, err)
				}
				if got != c.want {
					t.Errorf("%s = %q, want %q", c.query, got, c.want)
				}
			}
		})
	}
	if !ran {
		t.Skip("no Postgres, MySQL or MariaDB target: run with -alldbs or TESTDBS_ALL")
	}
}
