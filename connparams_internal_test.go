package testdbs

import "testing"

func TestServerDbName(t *testing.T) {
	tcs := []struct {
		name string
		in   string
		want string
	}{
		{name: "default keeps the case the container created it with", in: defaultDbName, want: "testdbDefault"},
		{name: "default in any case", in: "TESTDBDEFAULT", want: "testdbDefault"},
		{name: "other names are normalized", in: "TestFoo/Bar", want: "testfoobar"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if got := serverDbName(tc.in); got != tc.want {
				t.Errorf("serverDbName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
