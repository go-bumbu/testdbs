package testdbs

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"testing"
)

// resetInitState gives the test a package state where no init ran, and
// restores the state TestMain set up once the test ends. Tests using it must
// not run in parallel.
func resetInitState(t *testing.T) {
	t.Helper()
	savedDBs, savedInit := targetDBS, initialized
	targetDBS, initialized = []TargetDb{}, false
	t.Cleanup(func() { targetDBS, initialized = savedDBs, savedInit })
}

// fakeDb is a TargetDb with no database behind it. CloseAll records the call
// and returns closeErr.
type fakeDb struct {
	closeErr error
	closed   bool
}

func (f *fakeDb) DbType() string { return "fake" }

func (f *fakeDb) Init(logger.Interface) {}

func (f *fakeDb) Conn() *gorm.DB { return nil }

func (f *fakeDb) ConnDbName(string) *gorm.DB { return nil }

func (f *fakeDb) Close(string) error { return nil }

func (f *fakeDb) CloseAll() error {
	f.closed = true
	return f.closeErr
}

func TestDBsPanicsWithoutInit(t *testing.T) {
	resetInitState(t)
	if err := Clean(); err != nil {
		t.Errorf("Clean() before init = %v, want nil", err)
	}
	defer func() {
		if recover() == nil {
			t.Error("DBs() did not panic before init")
		}
	}()
	DBs()
}

func TestDBsEmptyWhenInitSelectsNothing(t *testing.T) {
	resetInitState(t)
	InitCustomDbs(nil, nil)

	if dbs := DBs(); len(dbs) != 0 {
		t.Fatalf("DBs() returned %d DBs, want 0", len(dbs))
	}
	if err := Clean(); err != nil {
		t.Errorf("Clean() = %v, want nil", err)
	}
}

func TestCleanClosesAllAndReturnsErrors(t *testing.T) {
	resetInitState(t)
	failing := &fakeDb{closeErr: errors.New("boom")}
	healthy := &fakeDb{}
	InitCustomDbs([]TargetDb{failing, healthy}, nil)

	err := Clean()
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("Clean() = %v, want an error mentioning boom", err)
	}
	if !failing.closed || !healthy.closed {
		t.Error("Clean() must close every DB, even after one fails")
	}
}

func TestNormalizeDbName(t *testing.T) {
	tcs := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain name", in: "custom", want: "custom"},
		{name: "subtest path", in: "TestStartDbNew/customDbTestName/SqliteNoCgo", want: "teststartdbnewcustomdbtestnamesqlitenocgo"},
		{name: "dots and spaces", in: "my db.name", want: "mydbname"},
		{name: "cut to 64 characters", in: strings.Repeat("a", 70), want: strings.Repeat("a", 64)},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeDbName(tc.in); got != tc.want {
				t.Errorf("normalizeDbName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
