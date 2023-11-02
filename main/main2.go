package main

import (
	"github.com/atmshang/rxfsnotify/fs"
	"log"
	"time"
)

func main() {
	fileSystem, err := fs.NewFileSystem("D:\\TempDir")
	if err != nil {
		log.Printf("Failed to build file system: %v\n", err)
		return
	}

	log.Println("==========1111111111111============")

	fileSystem.Print()

	fileSystem2, err := fs.NewFileSystem("D:\\TempDir")
	if err != nil {
		log.Printf("Failed to build file system: %v\n", err)
		return
	}

	log.Println("=========222222222222222=============")
	fileSystem2.Print()

	time.Sleep(10 * time.Second)

	err = fileSystem2.Update("D:\\TempDir\\待复制进来的文件夹\\图像")
	if err != nil {
		log.Printf("Failed to build file system: %v\n", err)
		return
	}

	log.Println("=========33333333=============")
	fileSystem2.Print()

	ret := fileSystem.Diff(fileSystem2)
	log.Println("======================")

	log.Println("[DIFF]", ret)
}
