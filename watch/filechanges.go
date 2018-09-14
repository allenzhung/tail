package watch

type FileChanges struct {
	Modified  chan bool
	Truncated chan bool
	Deleted   chan bool
}

func NewFileChanges() *FileChanges {
	return &FileChanges{
		make(chan bool, 1),
		make(chan bool, 1),
		make(chan bool, 1),
	}
}

func (fc *FileChanges) NotifyModified() {
	sendOnlyIfEmpty(fc.Modified)
}

func (fc *FileChanges) NotifyTruncated() {
	sendOnlyIfEmpty(fc.Truncated)
}

func (fc *FileChanges) NotifyDeleted() {
	sendOnlyIfEmpty(fc.Deleted)
}

func sendOnlyIfEmpty(ch chan bool) {
	select {
	case ch <- true:
	default:
	}
}
