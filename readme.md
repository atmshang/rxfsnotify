# rxfsnotify

一个基于fsnotify的文件系统事件监听库,使用rxgo实现事件过滤。

## 特性

- 使用fsnotify监听文件系统事件
- 支持监听多个目录
- 使用channel和rxgo过滤文件事件,避免重复处理
- 缓冲并批量处理文件事件
- 注册回调接口处理文件变化
- 并发安全的数据结构
- 优雅退出
- 可检测文件有效性
- 完整的示例程序

## 用法

```go
// 创建回调对象
type MyCallback struct{}

func (cb *MyCallback) OnPathChanged(cbe rxfsnotify.CallBackEvent) {
  // 处理路径变化事件  
  log.Println(cbe.Path, cbe.Exist) 
}

// 设置回调
callback := &MyCallback{}
rxfsnotify.SetPathCallbackListener(callback)

// 开始监听
rxfsnotify.Start([]string{"watch_dir_1", "watch_dir_2"}) 

// 优雅停止
rxfsnotify.GracefulStop()
```

## 运行示例

`go run main/main.go`

## 贡献

欢迎提issue和PR来贡献代码!

## License

rxfsnotify is released under the [MIT License](https://opensource.org/licenses/MIT).