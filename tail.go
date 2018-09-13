package tail

import (
	"strings"
	"sync"
	"bufio"
	"os"
	"time"
	"io"
)

type Line struct {
	Text string
	Time time.Time
	Err  error // Error from tail
}

// NewLine returns a Line with present time.
func NewLine(text string) *Line {
	return &Line{text, time.Now(), nil}
}

type Tail struct {
	Filename 	string
	Lines		chan *Line

	Follow		bool


	file 		*os.File
	reader 		*bufio.Reader
	lk sync.Mutex
}

func TailFile(filename string)(*Tail, error) {
	t := &Tail {
		Filename:	filename,
		Lines:		make(chan *Line),
	}

	return t, nil
}


func (tail *Tail) tailFileSync(){
	tail.openReader()
	for {
		line, err := tail.readLine()
		if err == nil {
			tail.sendLine(line)
		} else if err == io.EOF {
			if !tail.Follow {
				if line != "" {
					tail.sendLine(line)
				}
				return
			}
			if tail.Follow && line != "" {
				// to do
				return
			}


		} else {
			// 既不是文件结尾，也没有error
			return
		}


		select {
			// todo
		}
	}
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










