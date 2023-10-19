package rxfsnotify

import (
	"os"
	"time"
)

// 判断文件是否有效
func isValidFile(filePath string) bool {
	// os.Stat 获取 filePath 的文件信息
	_, err := os.Stat(filePath)
	// 如果在获取文件信息过程中出现错误（比如文件不存在，路径错误等），则返回 false
	if err != nil {
		return false
	}
	// 若 filePath 是一个存在的文件，返回 true
	return true
}

// 判断文件是否处于空闲状态
func isIdleFile(filePath string) bool {
	noNeedFile, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer noNeedFile.Close()
	return true
}

func checkFileUntilValidOrIdle(filePath string) bool {
	for {
		// 先等待一会儿
		time.Sleep(250 * time.Millisecond)

		// 检查文件是否有效
		if !isValidFile(filePath) {
			//plog.Println("Check: IsNotValidFile", filePath)
			return false
		}
		// 检查文件是否空闲
		if !isIdleFile(filePath) {
			//plog.Println("Wait: IsNotIdleFile", filePath)
			continue
		}
		return true
	}
}
