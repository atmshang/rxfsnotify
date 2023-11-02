package main

import (
	"github.com/atmshang/plog"
	"github.com/atmshang/rxfsnotify"
	"log"
)

type MyCallback struct{}

func (cb *MyCallback) OnPathChanged(cbe rxfsnotify.CallBackEvent) {
	// 处理路径变化事件
	log.Println("[SDK的回调] ", cbe.Path, cbe.Exist)
}

func main() {
	plog.Println("Start")
	plog.SetEnable(false)

	callback := &MyCallback{}

	plog.Println("SetPathCallbackListener")
	rxfsnotify.SetPathCallbackListener(callback)

	go rxfsnotify.Start([]string{"D:\\TempDir"})

	select {}
}
