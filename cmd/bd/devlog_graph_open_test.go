package main

import (
	"reflect"
	"testing"
)

// TestBrowserOpenCommand pins the per-platform opener argv so --open uses the
// right launcher on each OS (verified without actually spawning anything).
func TestBrowserOpenCommand(t *testing.T) {
	cases := []struct {
		goos     string
		wantName string
		wantArgs []string
	}{
		{"darwin", "open", []string{"/tmp/g.html"}},
		{"linux", "xdg-open", []string{"/tmp/g.html"}},
		{"freebsd", "xdg-open", []string{"/tmp/g.html"}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", "/tmp/g.html"}},
	}
	for _, c := range cases {
		name, args := browserOpenCommand(c.goos, "/tmp/g.html")
		if name != c.wantName || !reflect.DeepEqual(args, c.wantArgs) {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", c.goos, name, args, c.wantName, c.wantArgs)
		}
	}
}
