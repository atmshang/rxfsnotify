package rxfsnotify

import (
	"github.com/atmshang/plog"
	"github.com/reactivex/rxgo/v2"
	"sync"
	"time"
)

var fileEventCh = make(chan rxgo.Item)

// 发送事件到管道的方法，闻到味了.jpg
func sendFileEvent(info fileEvent) {
	defer func() {
		if r := recover(); r != nil {
			plog.Println("recover:", r)
		}
	}()
	fileEventCh <- rxgo.Item{V: info, E: nil}
}

func tryCloseFileEventCh() {
	defer func() {
		if r := recover(); r != nil {
			plog.Println("recover:", r)
		}
	}()
	close(fileEventCh)
}

var eventFilterLocker sync.Mutex

func fileFilter() {
	tryCloseFileEventCh()

	eventFilterLocker.Lock()
	defer eventFilterLocker.Unlock()

	fileEventCh = make(chan rxgo.Item)
	observable := rxgo.FromChannel(fileEventCh).
		BufferWithTimeOrCount(rxgo.WithDuration(time.Millisecond*250), 5).
		FlatMap(func(item rxgo.Item) rxgo.Observable {
			return rxgo.Just(item.V)()
		})
	for item := range observable.Observe() {
		info, ok := item.V.(fileEvent)
		if ok {
			filePath := info.Path
			plog.Println("接收事件：", info)
			go dealWithFileEvent(filePath)
		}
	}
}

var fileLocks = make(map[string]*sync.Mutex)

func dealWithFileEvent(filePath string) {
	// 检查锁
	lock, ok := fileLocks[filePath]
	if !ok {
		lock = &sync.Mutex{}
		fileLocks[filePath] = lock
	}

	ok = lock.TryLock()
	if !ok {
		plog.Println("跳过事件：", filePath)
		return
	}
	defer lock.Unlock()
	plog.Println("处理事件：", filePath)

	// 会被阻塞在检查中
	valid := checkFileUntilValidOrIdle(filePath)
	if !valid {
		plog.Println("结果: File is not exist:", filePath)
	} else {
		plog.Println("结果: File is changed:", filePath)
	}

	go callback(filePath, valid)

}

func callback(filePath string, exist bool) {
	if cb != nil {
		cbe := CallBackEvent{Path: filePath, Exist: exist}
		defer func() {
			if r := recover(); r != nil {
				plog.Println("callback recover:", r)
			}
		}()
		cb.OnPathChanged(cbe)
	}
}
