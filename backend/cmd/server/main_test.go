package main

import "testing"

func TestVersionRequested(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"version"}} {
		if !versionRequested(args) {
			t.Fatalf("versionRequested(%v) = false", args)
		}
	}
	for _, args := range [][]string{nil, {"serve"}, {"--version", "extra"}} {
		if versionRequested(args) {
			t.Fatalf("versionRequested(%v) = true", args)
		}
	}
}
