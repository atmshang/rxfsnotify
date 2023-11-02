package concurrent

import (
	"github.com/atmshang/plog"
	"sync"
	"time"
)

type Task struct {
	ID       uint64
	Delay    time.Duration
	Canceled bool
	Execute  func()
}

type TaskQueue struct {
	mu        sync.Mutex
	tasks     []*Task
	counter   uint64
	newTask   chan *Task
	isRunning bool
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		newTask: make(chan *Task),
	}
}

func (tq *TaskQueue) AddTask(delay time.Duration, execute func()) *Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task := &Task{
		ID:      tq.counter,
		Delay:   delay,
		Execute: execute,
	}

	tq.tasks = append(tq.tasks, task)

	// Check if the counter has reached the maximum value of uint64.
	// If so, reset it to 0. Otherwise, increment it.
	if tq.counter == ^uint64(0) { // ^uint64(0) gives the maximum value of uint64
		tq.counter = 0
	} else {
		tq.counter++
	}

	tq.newTask <- task

	return task
}

func (tq *TaskQueue) CancelAll() {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	for _, task := range tq.tasks {
		task.Canceled = true
	}

	tq.tasks = nil
	tq.isRunning = false
}

func (tq *TaskQueue) Start() {
	tq.mu.Lock()
	if tq.isRunning {
		tq.mu.Unlock()
		return
	}
	tq.isRunning = true
	tq.mu.Unlock()

	tq.CancelAll()

	go func() {
		for {
			task := <-tq.newTask
			// Start a new Goroutine for each new task.
			plog.Println("未来执行新任务：", task.ID)
			go func(task *Task) {
				select {
				case <-time.After(task.Delay):
					plog.Println("执行任务：", task.ID)
					tq.executeTask(task)
					plog.Println("执行完成：", task.ID)
				}
			}(task)
		}
	}()
}

func (tq *TaskQueue) getNextTask() *Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.tasks) == 0 {
		return nil
	}

	task := tq.tasks[0]
	tq.tasks = tq.tasks[1:]

	return task
}

func (tq *TaskQueue) executeTask(task *Task) {
	if task.Canceled {
		plog.Println("任务执行前被取消：", task.ID)
		return
	}
	plog.Println("任务内容被执行：", task.ID)
	task.Execute()
	plog.Println("任务内容执行完成", task.ID)
}
