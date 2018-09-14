package watch

import "gopkg.in/tomb.v1"


type FileWatcher interface {
	BlockUntilExists(*tomb.Tomb) error

	ChangeEvents(*tomb.Tomb, int64) (*FileChanges, error)
}