// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package fsnotify

import (
	"os"
	"testing"
	"time"
	"vendor"
)

func TestEventStringWithValue(t *testing.T) {
	for opMask, expectedString := range map[vendor.Op]string{
		vendor.Chmod | vendor.Create: `"/usr/someFile": CREATE|CHMOD`,
		vendor.Rename:                `"/usr/someFile": RENAME`,
		vendor.Remove:                `"/usr/someFile": REMOVE`,
		vendor.Write | vendor.Chmod:  `"/usr/someFile": WRITE|CHMOD`,
	} {
		event := vendor.Event{Name: "/usr/someFile", Op: opMask}
		if event.String() != expectedString {
			t.Fatalf("Expected %s, got: %v", expectedString, event.String())
		}

	}
}

func TestEventOpStringWithValue(t *testing.T) {
	expectedOpString := "WRITE|CHMOD"
	event := vendor.Event{Name: "someFile", Op: vendor.Write | vendor.Chmod}
	if event.Op.String() != expectedOpString {
		t.Fatalf("Expected %s, got: %v", expectedOpString, event.Op.String())
	}
}

func TestEventOpStringWithNoValue(t *testing.T) {
	expectedOpString := ""
	event := vendor.Event{Name: "testFile", Op: 0}
	if event.Op.String() != expectedOpString {
		t.Fatalf("Expected %s, got: %v", expectedOpString, event.Op.String())
	}
}

// TestWatcherClose tests that the goroutine started by creating the watcher can be
// signalled to return at any time, even if there is no goroutine listening on the events
// or errors channels.
func TestWatcherClose(t *testing.T) {
	t.Parallel()

	name := vendor.tempMkFile(t, "")
	w := vendor.newWatcher(t)
	err := w.Add(name)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
	// Allow the watcher to receive the event.
	time.Sleep(time.Millisecond * 100)

	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
}
