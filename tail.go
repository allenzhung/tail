package tail

import (
	"fmt"
	"gopkg.in/tomb.v1"
	"strings"
	"sync"
	"bufio"
	"os"
	"time"
	"io"
	"tail/watch"
	"errors"
)

var (
	ErrStop = errors.New("tail should now stop")
)

type Line struct {
	Text string
	Time time.Time
	Err  error // Error from tail
}

type SeekInfo struct {
	Offset 		int64
	Whence		int	
}

// NewLine returns a Line with present time.
func NewLine(text string) *Line {
	return &Line{text, time.Now(), nil}
}

type Tail struct {
	Filename 	string
	Lines		chan *Line

	Follow		bool

	watcher 	watch.FileWatcher
	changes 	*watch.FileChanges

	tomb.Tomb	

	file 		*os.File
	reader 		*bufio.Reader
	lk sync.Mutex
}


func TailFile(filename string)(*Tail, error) {
	t := &Tail {
		Filename:	filename,
		Lines:		make(chan *Line),
	}
	go t.tailFileSync()
	return t, nil
}


func (tail *Tail) tailFileSync(){
	tail.openReader()

	var offset int64
	var err error

	for {
		line, err := tail.readLine()
		if err == nil {
			tail.sendLine(line)
		} else if err == io.EOF {
			// 表示读到文件的最后了
			if !tail.Follow {
				if line != "" {
					tail.sendLine(line)
				}
				return
			}
			if tail.Follow && line != "" {
				err := tail.seekTo(SeekInfo{Offset: offset, Whence: 0})
				if err != nil {
					tail.Kill(err)
					return
				}
			}

			err := tail.waitForChanges()
			if err != nil {
				if err != ErrStop {
					tail.Kill(err)
				}
				return
			}


		} else {
			// 既不是文件结尾，也没有error
			return
		}


		select {
			//TODO: 未完成、
		}
	}
}

func (tail *Tail) waitForChanges() error {
	if tail.changes == nil {
		// 这里是获取文件指针的当前位置
		pos, err := tail.file.Seek(0,os.SEEK_CUR)
		if err != nil {
			return err
		}
		tail.changes, err = tail.watcher.ChangeEvents(&tail.Tomb, pos)
		if err != nil {
			return err
		}
	}
	select {
	case <- tail.changes.Modified:
		return nil
	case <- tail.changes.Deleted:
		tail.changes = nil
		// TODO: 为完成
	case <- tail.changes.Truncated:
		// TODO: 需要补充
		tail.openReader()
		return nil
	case <- tail.Dying():
		return nil
	}
	panic("unreachable")
}


func (tail *Tail) openReader(){
	tail.reader = bufio.NewReader(tail.file)
}

func (tail *Tail) readLine()(string,error) {
	tail.lk.Lock()
	line, err := tail.reader.ReadString('\n')
	tail.lk.Unlock()

	if err != nil {
		return line, err
	}
	line = strings.TrimRight(line, "\n")
	return line, err
}

func (tail *Tail) sendLine(line string) bool {
	now := time.Now()
	lines := []string{line}

	for _, line := range lines {
		tail.Lines <- &Line {
			line,
			now,
			nil,
		}
	}
	return true
}


func (tail *Tail) seekTo(pos SeekInfo) error {
	_, err := tail.file.Seek(pos.Offset, pos.Whence)
	if err != nil {
		return fmt.Errorf("seek error on %s: %s", tail.Filename, err)
	}
	tail.reader.Reset(tail.file)
	return nil
}

func (tail *Tail) seekEnd() error {
	return tail.seekTo(SeekInfo{Offset:0,Whence: os.SEEK_CUR})
}












