package rxfsnotify

import (
	"github.com/atmshang/plog"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var stopCh = make(chan bool)
var wg sync.WaitGroup

func Start(dirPaths []string) {
	tryCloseCh()
	stopCh = make(chan bool)
	wg.Add(1)
	run(dirPaths)
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

	plog.Println("添加观察目录：", dirPath)
	err := watcher.Add(dirPath) //添加观察目录
	if err != nil {
		// 忽略这个情况即可
	}
}

func eventHandler(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			filePath := event.Name

			// 持续添加新的被观察目录
			if event.Op == fsnotify.Create {
				if stat, err := os.Stat(filePath); err == nil && stat.IsDir() {
					// 递归补充添加新的子目录
					refreshWatchedPaths(watcher, []string{filePath})
				}
			}

			//其他监听流程

			op := event.Op.String()

			_event := fileEvent{Path: filePath, Event: op}
			plog.Println("发送事件:", _event)
			sendFileEvent(_event)

		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			plog.Println("Ignore err:", err)

		case <-stopCh: // 当接收到来自GracefulStop的信号时，结束循环
			plog.Println("eventHandler结束循环")
			// 通知所有协程
			tryCloseCh()
			return
		}
	}
}

func tryCloseCh() {
	defer func() {
		if r := recover(); r != nil {
			plog.Println("recover:", r)
		}
	}()
	close(stopCh)
}
