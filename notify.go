package rxfsnotify

import (
	"github.com/atmshang/plog"
	"github.com/atmshang/rxfsnotify/concurrent"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

var stopCh = make(chan bool)
var wg sync.WaitGroup

func Start(dirPaths []string) {
	tryCloseCh()
	refreshTaskQueue.Start()
	stopCh = make(chan bool)
	wg.Add(1)
	run(dirPaths)
	refreshTaskQueue.CancelAll()
}

func GracefulStop() {
	stopCh <- true
	wg.Wait()
}

var singleLocker sync.Mutex

// 这个方法只能单例运行
func run(dirPaths []string) {
	defer wg.Done()

	singleLocker.Lock()
	defer singleLocker.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		plog.Panic(err)
	}
	defer watcher.Close()

	refreshWatchedPaths(watcher, dirPaths) //持续添加目录

	go fileFilter() //启动过滤器

	go eventHandler(watcher) //注册观察回调

	<-stopCh // 这里会阻塞，直到接收到来自GracefulStop的信号
	plog.Println("主协程退出")
	// 通知所有协程退出
	tryCloseCh()
}

// 开始监视路径的变化。
// 一个路径只能被监视一次；尝试多次监视同一路径将返回错误。尚不存在于文件系统上的路径无法被添加监视。如果路径被删除，监视将自动删除。
// 如果将路径重命名到同一文件系统上的其他位置，监视将保持不变，但是如果路径被删除并重新创建，或者被移动到不同的文件系统，监视将被删除。
// 在网络文件系统（NFS、SMB、FUSE 等）或特殊文件系统（/proc、/sys 等）上通常无法正常工作。
// 所有目录中的文件都将被监视，包括在观察器启动后创建的新文件。子目录不会被监视（即非递归）。
// 通常不建议仅监视单个文件（而不是目录），因为许多工具以原子方式更新文件。而不是直接写入文件，首先会写入临时文件，如果成功，则将临时文件移动到目标位置，删除原始文件，或者进行某种变体。原始文件上的监视器现在丢失了，因为它不再存在。
// 相反，监视父目录并使用 Event.Name 过滤您不感兴趣的文件。在 [cmd/fsnotify/file.go] 中有一个示例。
func refreshWatchedPaths(watcher *fsnotify.Watcher, dirPaths []string) {

	watchedPaths := make(map[string]bool)

	for _, dirPath := range dirPaths {
		traverseDir(watchedPaths, dirPath)
	}
	var finalPaths []string
	for p := range watchedPaths {
		finalPaths = append(finalPaths, p)
	}

	for _, dirPath := range finalPaths {
		addWatchedPaths(watcher, dirPath)
	}
}

func traverseDir(watchedPaths map[string]bool, dirPath string) {
	watchedPaths[dirPath] = true

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		plog.Println(err)
		return
	}

	for _, f := range files {
		fp := filepath.Join(dirPath, f.Name())
		if f.IsDir() {
			watchedPaths[fp] = true
			traverseDir(watchedPaths, fp)
		} else {
			// 不需要管文件
		}
	}
}

var addLocker sync.Mutex

func addWatchedPaths(watcher *fsnotify.Watcher, dirPath string) {

	addLocker.Lock()
	defer addLocker.Unlock()

	err := watcher.Remove(dirPath)
	if err != nil {
		log.Println("[REMOVE4ADD] 移除观察失败：", err, dirPath)
	} else {
		log.Println("[REMOVE4ADD] 移除观察成功：", dirPath)
	}

	err = watcher.Add(dirPath) //添加观察目录
	if err != nil {
		log.Println("[ADD] 添加观察目录失败：", err, dirPath)
		return
	}
	log.Println("[ADD] 添加观察目录成功：", dirPath)
}

// fsnotify的文档有说，会自动移除不存在的监听，先不管他，但是rename的情况有bug，难顶
func removeWatch(watcher *fsnotify.Watcher, event fsnotify.Event) {
	addLocker.Lock()
	defer addLocker.Unlock()

	err := watcher.Remove(event.Name)
	if err != nil {
		log.Println("[REMOVE] 移除观察失败：", err, event.Name)
		return
	}
	log.Println("[REMOVE] 移除观察成功：", event.Name)
}

var waitingRefreshDirMap = concurrent.NewSafeMap()

var refreshTaskQueue = concurrent.NewTaskQueue()

func innerNotify(event fsnotify.Event) {
	_event := fileEvent{Path: event.Name, Event: event.Op.String()}
	sendFileEvent(_event)

}

func innerProcessDir(watcher *fsnotify.Watcher, event fsnotify.Event) {
	// 标记变更
	waitingRefreshDirMap.Set(event.Name, true)
	// 取消等待的任务
	refreshTaskQueue.CancelAll()
	// 发布新任务到未来
	_ = refreshTaskQueue.AddTask(5000*time.Millisecond, func() {
		dirPaths := waitingRefreshDirMap.ToList()
		for _, dirPath := range dirPaths {
			_event := fileEvent{Path: dirPath, Event: fsnotify.Create.String()}
			plog.Println("批量发送文件夹新建的事件:", _event)
			sendFileEvent(_event)
		}
		refreshWatchedPaths(watcher, dirPaths)
	})

}

func eventHandler(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				plog.Println("监听事件管道发现：不OK")
				continue
			}
			// 判断状态
			stat, err := os.Stat(event.Name)
			if err != nil {
				plog.Println("这个文件不在啦：", event)
				removeWatch(watcher, event)
				innerNotify(event)
				continue
			} else {
				plog.Println("这个文件还在：", event)
			}
			if stat.IsDir() {
				switch event.Op {
				case fsnotify.Create:
					plog.Println("这个文件是个新建的文件夹：", event)
					innerProcessDir(watcher, event)
					continue
				case fsnotify.Write:
					plog.Println("这个文件是个有内容变化的文件夹：", event)
				case fsnotify.Remove:
					// remove分为要remove和已经remove两次，要remove这回事，我们不管，这里直接continue，这样会走上面文件不在了的流程
					plog.Println("这个文件是被移除的文件夹：", event)
					removeWatch(watcher, event)
					continue
				case fsnotify.Rename:
					plog.Println("这个文件是被重命名的文件夹：", event)
					removeWatch(watcher, event)
					continue
				case fsnotify.Chmod:
					plog.Println("这个文件是被修改权限的文件夹：", event)
				default:
					plog.Println("未知操作：", event)
				}
			} else {
				plog.Println("这个文件是文件：", event)
				innerNotify(event)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				plog.Println("监听错误管道发现：不OK")
				continue
			}
			plog.Println("监听错误管道发现:", err)

		case <-stopCh: // 当接收到来自GracefulStop的信号时，结束循环
			plog.Println("决定优雅退出")
			// 通知所有协程
			tryCloseCh()
			return
		}
	}
}

func tryCloseCh() {
	defer func() {
		if r := recover(); r != nil {
			// plog.Println("recover:", r)
			debug.PrintStack()
		}
	}()
	close(stopCh)
}
