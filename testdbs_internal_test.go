package testdbs

import "testing"

// resetInitState gives the test a package state where no init ran, and
// restores the state TestMain set up once the test ends. Tests using it must
// not run in parallel.
func resetInitState(t *testing.T) {
	t.Helper()
	savedDBs, savedInit := targetDBS, initialized
	targetDBS, initialized = []TargetDb{}, false
	t.Cleanup(func() { targetDBS, initialized = savedDBs, savedInit })
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
