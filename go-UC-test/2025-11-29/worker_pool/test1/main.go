package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	taskCnt = 10
)

func DoFuncByWorkerPool(workerCnt int) {
	var wg sync.WaitGroup

	taskChan := make(chan int, taskCnt)
	resChan := make(chan int, taskCnt)

	for i := 0; i < workerCnt; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for task := range taskChan {
				fmt.Printf("Worker %d: 处理任务 %d\n", workerId, task)
				resChan <- task * task
			}
		}(i)
	}

	for i := range taskCnt {
		taskChan <- i
	}
	close(taskChan)

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for item := range resChan {
		fmt.Printf("收到结果: %d\n", item)
	}
}

func testChan() {
	Tchan := make(chan int)

	go func() {
		value, ok := <-Tchan
		if !ok {
			fmt.Println("channel is closed")
			return
		}
		fmt.Printf("value is %d", value)
	}()

	Tchan <- 1
	time.Sleep(2 * time.Second)
}

func main() {
	// DoFuncByWorkerPool(10)
	testChan()
}
