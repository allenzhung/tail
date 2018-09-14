package watch

import (
	"os"
	"gopkg.in/tomb.v1"
	"path/filepath"
	"gopkg.in/fsnotify.v1"
)


type InotifyFileWatcher struct {
	Filename 	string
	Size		int64
}


func NewInotifyFileWatcher(filename string) *InotifyFileWatcher {
	fw := &InotifyFileWatcher {
		filepath.Clean(filename),
		0,
	}
	return fw
}

func (fw *InotifyFileWatcher) ChangeEvents(t *tomb.Tomb, pos int64) (*FileChanges, error) {
	err := Watch(fw.Filename)
	if err != nil {
		return nil, err
	}

	changes := NewFileChanges()
	fw.Size = pos

	go func() {
		events := Events(fw.Filename)

		for {
			prevSize := fw.Size

			var evt fsnotify.Event
			var ok bool
			
			select {
			case evt, ok = <- events:
				if !ok {
					RemoveWatch(fw.Filename)
					return
				}
			case <- t.Dying():
				RemoveWatch(fw.Filename)
				return
			}
			switch {
			case evt.Op & fsnotify.Remove == fsnotify.Remove:
				fallthrough
			case evt.Op & fsnotify.Rename == fsnotify.Rename:
				RemoveWatch(fw.Filename)
				changes.NotifyDeleted()
				return

			case evt.Op & fsnotify.Chmod == fsnotify.Chmod:
				fallthrough
			case evt.Op & fsnotify.Write == fsnotify.Write:
				fi, err := os.Stat(fw.Filename)
				if err != nil {
					if os.IsNotExist(err) {
						RemoveWatch(fw.Filename)
						changes.NotifyDeleted()
						return
					}

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
