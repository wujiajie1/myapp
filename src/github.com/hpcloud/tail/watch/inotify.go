// Copyright (c) 2015 HPE Software Inc. All rights reserved.
// Copyright (c) 2013 ActiveState Software Inc. All rights reserved.

package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"vendor"
)

// InotifyFileWatcher uses inotify to monitor file changes.
type InotifyFileWatcher struct {
	Filename string
	Size     int64
}

func NewInotifyFileWatcher(filename string) *InotifyFileWatcher {
	fw := &InotifyFileWatcher{filepath.Clean(filename), 0}
	return fw
}

func (fw *InotifyFileWatcher) BlockUntilExists(t *vendor.Tomb) error {
	err := vendor.WatchCreate(fw.Filename)
	if err != nil {
		return err
	}
	defer vendor.RemoveWatchCreate(fw.Filename)

	// Do a real check now as the file might have been created before
	// calling `WatchFlags` above.
	if _, err = os.Stat(fw.Filename); !os.IsNotExist(err) {
		// file exists, or stat returned an error.
		return err
	}

	events := vendor.Events(fw.Filename)

	for {
		select {
		case evt, ok := <-events:
			if !ok {
				return fmt.Errorf("inotify watcher has been closed")
			}
			evtName, err := filepath.Abs(evt.Name)
			if err != nil {
				return err
			}
			fwFilename, err := filepath.Abs(fw.Filename)
			if err != nil {
				return err
			}
			if evtName == fwFilename {
				return nil
			}
		case <-t.Dying():
			return vendor.ErrDying
		}
	}
	panic("unreachable")
}

func (fw *InotifyFileWatcher) ChangeEvents(t *vendor.Tomb, pos int64) (*vendor.FileChanges, error) {
	err := vendor.Watch(fw.Filename)
	if err != nil {
		return nil, err
	}

	changes := vendor.NewFileChanges()
	fw.Size = pos

	go func() {

		events := vendor.Events(fw.Filename)

		for {
			prevSize := fw.Size

			var evt vendor.Event
			var ok bool

			select {
			case evt, ok = <-events:
				if !ok {
					vendor.RemoveWatch(fw.Filename)
					return
				}
			case <-t.Dying():
				vendor.RemoveWatch(fw.Filename)
				return
			}

			switch {
			case evt.Op&vendor.Remove == vendor.Remove:
				fallthrough

			case evt.Op&vendor.Rename == vendor.Rename:
				vendor.RemoveWatch(fw.Filename)
				changes.NotifyDeleted()
				return

			//With an open fd, unlink(fd) - inotify returns IN_ATTRIB (==fsnotify.Chmod)
			case evt.Op&vendor.Chmod == vendor.Chmod:
				fallthrough

			case evt.Op&vendor.Write == vendor.Write:
				fi, err := os.Stat(fw.Filename)
				if err != nil {
					if os.IsNotExist(err) {
						vendor.RemoveWatch(fw.Filename)
						changes.NotifyDeleted()
						return
					}
					// XXX: report this error back to the user
					vendor.Fatal("Failed to stat file %v: %v", fw.Filename, err)
				}
				fw.Size = fi.Size()

				if prevSize > 0 && prevSize > fw.Size {
					changes.NotifyTruncated()
				} else {
					changes.NotifyModified()
				}
				prevSize = fw.Size
			}
		}
	}()

	return changes, nil
}
