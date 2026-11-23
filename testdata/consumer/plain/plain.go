// Package plain does not import testdbs, like most packages of a user's
// module. `TESTDBS_ALL=1 go test ./...` must still pass here, where the
// -alldbs flag would fail as undefined.
package plain

func Double(n int) int { return 2 * n }
