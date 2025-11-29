package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	wokerCount = 3
	taskCount  = 10
)

func taskFunc() {
	var wg sync.WaitGroup

	// 任务和结果通道
	taskChan := make(chan int, taskCount)
	resultChan := make(chan int, taskCount)

	for i := 0; i <= wokerCount; i++ {
		wg.Add(1)
		go func(wokerID int) {
			defer wg.Done()
			for task := range taskChan {
				fmt.Printf("Worker %d: 处理任务 %d\n", wokerID, task)
				time.Sleep(50 * time.Millisecond)
				resultChan <- task * 2
			}
		}(i)
	}

	// 向任务通道中发送数据的时候，发送完毕需要关闭通道
	// 否则遍历通道的主goroutine会一直阻塞等待新元素
	for i := range taskCount {
		taskChan <- i
	}
	close(taskChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		fmt.Printf("收到结果: %d\n", res)
	}

}

func main() {
	taskFunc()
}
