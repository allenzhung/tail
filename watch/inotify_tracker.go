package watch

import (
	"os"
	"path/filepath"
	"sync"
	"gopkg.in/fsnotify.v1"
	"syscall"
)


type watchInfo struct {
	op 		fsnotify.Op
	fname 	string
}

func (this *watchInfo) isCreate() bool {
	return this.op == fsnotify.Create
}

type InotifyTracker struct {
	mux 		sync.Mutex
	watcher		*fsnotify.Watcher
	chans		map[string] chan fsnotify.Event
	done		map[string] chan bool
	watchNums	map[string] int
	watch		chan *watchInfo
	remove 		chan *watchInfo
	error		chan error
}


var (
	shared *InotifyTracker

	once = sync.Once{}
	goRun = func() {
		shared = &InotifyTracker {
			mux:		sync.Mutex{},
			chans:		make(map[string] chan fsnotify.Event),
			done:		make(map[string] chan bool),
			watchNums:	make(map[string]int),
			watch:		make(chan *watchInfo),
			remove:		make(chan *watchInfo),
			error:		make(chan error),
		}
		//go shared.run()
	}
)

func Watch(fname string) error {
	return watch (&watchInfo {
		fname: fname,
	})
}

func WatchCreate(fname string) error {
	return watch(&watchInfo {
		op:		fsnotify.Create,
		fname:	fname,
	})
}

func watch(winfo *watchInfo) error {
	once.Do(goRun)

	winfo.fname = filepath.Clean(winfo.fname)
	shared.watch <- winfo
	return <- shared.error
}

func RemoveWatch(fname string) error {
	return remove(&watchInfo {
		fname: fname,
	})
}

func RemoveWatchCreate(fname string)error {
	return remove (&watchInfo {
		op:		fsnotify.Create,
		fname:	fname,
	})
}

func remove(winfo *watchInfo) error {
	once.Do(goRun)
	winfo.fname = filepath.Clean(winfo.fname)
	shared.mux.Lock()
	done := shared.done[winfo.fname]
	if done != nil {
		delete(shared.done, winfo.fname)
		close(done)
	}
	shared.mux.Unlock()

	shared.remove <- winfo
	return <- shared.error
}


func Events(fname string) <-chan fsnotify.Event {
	shared.mux.Lock()
	defer shared.mux.Unlock()

	return shared.chans[fname]
}

func Cleanup(fname string) error {
	return RemoveWatch(fname)
}


func (shared *InotifyTracker) addWatch(winfo *watchInfo) error {
	shared.mux.Lock()
	defer shared.mux.Unlock()

	if shared.chans[winfo.fname] == nil {
		shared.chans[winfo.fname] = make(chan fsnotify.Event)
	}

	if shared.done[winfo.fname] == nil {
		shared.done[winfo.fname] = make(chan bool)
	}

	fname := winfo.fname
	if winfo.isCreate() {
		fname = filepath.Dir(fname)
	}

	var err error

	if shared.watchNums[fname] == 0 {
		err = shared.watcher.Add(fname)
	}
	if err == nil {
		shared.watchNums[fname] ++
	}
	return err
}


func (shared *InotifyTracker) removeWatch(winfo *watchInfo) error {
	shared.mux.Lock()

	ch := shared.chans[winfo.fname]

	if ch!= nil {
		delete(shared.chans, winfo.fname)
		close(ch)
	}

	fname := winfo.fname
	if winfo.isCreate(){
		fname = filepath.Dir(fname)
	}

	shared.watchNums[fname]--
	watchNum := shared.watchNums[fname]
	if watchNum == 0 {
		delete(shared.watchNums,fname)
	}
	shared.mux.Unlock()

	var err error 
	if watchNum == 0 {
		err = shared.watcher.Remove(fname)
	}
	return err
}


func(shared *InotifyTracker) sendEvent(event fsnotify.Event) {
	name := filepath.Clean(event.Name)

	shared.mux.Lock()
	ch := shared.chans[name]
	done := shared.done[name]
	shared.mux.Unlock()

	if ch != nil && done != nil {
		select {
		case ch <- event:
		case <- done:
		}
	}
}


func (shared *InotifyTracker) run(){
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		//
	}
	shared.watcher = watcher

	for {
		select {
		case winfo := <-shared.watch:
			shared.error <- shared.addWatch(winfo)
		case winfo := <- shared.remove:
			shared.error <- shared.removeWatch(winfo)
		case event, open := <- shared.watcher.Events:
			if !open{
				return
			}
			shared.sendEvent(event)
		case err,open := <- shared.watcher.Errors:
			if !open {
				return
			} else if err != nil {
				sysErr, ok := err.(*os.SyscallError)
				if !ok || sysErr.Err != syscall.EINTR{
					// to do

				}
			}
		}
	}
}





