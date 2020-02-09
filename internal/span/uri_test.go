// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package span_test

import (
	"testing"

	"github.com/nokia/ntt/internal/span"
)

// TestURI tests the conversion between URIs and filenames. The test cases
// include Windows-style URIs and filepaths, but we avoid having OS-specific
// tests by using only forward slashes, assuming that the standard library
// functions filepath.ToSlash and filepath.FromSlash do not need testing.
func TestURI(t *testing.T) {
	for _, test := range []struct {
		path, wantFile string
		wantURI        span.URI
	}{
		{
			path:     ``,
			wantFile: ``,
			wantURI:  span.URI(""),
		},
		{
			path:     `C:/Windows/System32`,
			wantFile: `C:/Windows/System32`,
			wantURI:  span.URI("file:///C:/Windows/System32"),
		},
		{
			path:     `C:/Go/src/bob.go`,
			wantFile: `C:/Go/src/bob.go`,
			wantURI:  span.URI("file:///C:/Go/src/bob.go"),
		},
		{
			path:     `c:/Go/src/bob.go`,
			wantFile: `C:/Go/src/bob.go`,
			wantURI:  span.URI("file:///C:/Go/src/bob.go"),
		},
		{
			path:     `/path/to/dir`,
			wantFile: `/path/to/dir`,
			wantURI:  span.URI("file:///path/to/dir"),
		},
		{
			path:     `/a/b/c/src/bob.go`,
			wantFile: `/a/b/c/src/bob.go`,
			wantURI:  span.URI("file:///a/b/c/src/bob.go"),
		},
		{
			path:     `c:/Go/src/bob george/george/george.go`,
			wantFile: `C:/Go/src/bob george/george/george.go`,
			wantURI:  span.URI("file:///C:/Go/src/bob%20george/george/george.go"),
		},
		{
			path:     `file:///c:/Go/src/bob%20george/george/george.go`,
			wantFile: `C:/Go/src/bob george/george/george.go`,
			wantURI:  span.URI("file:///C:/Go/src/bob%20george/george/george.go"),
		},
		{
			path:     `file:///C%3A/Go/src/bob%20george/george/george.go`,
			wantFile: `C:/Go/src/bob george/george/george.go`,
			wantURI:  span.URI("file:///C:/Go/src/bob%20george/george/george.go"),
		},
		{
			path:     `file:///path/to/%25p%25ercent%25/per%25cent.go`,
			wantFile: `/path/to/%p%ercent%/per%cent.go`,
			wantURI:  span.URI(`file:///path/to/%25p%25ercent%25/per%25cent.go`),
		},
	} {
		got := span.NewURI(test.path)
		if got != test.wantURI {
			t.Errorf("NewURI(%q): got %q, expected %q", test.path, got, test.wantURI)
		}
		gotFilename := got.Filename()
		if gotFilename != test.wantFile {
			t.Errorf("Filename(%q): got %q, expected %q", got, gotFilename, test.wantFile)
		}
	}
}
