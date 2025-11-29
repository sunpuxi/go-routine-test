package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================
// 演示：Goroutine 的工作原理
// ============================================

// 演示1：Goroutine 的创建和执行
func demo1_BasicGoroutine() {
	fmt.Println("=== 演示1：基本 Goroutine 创建 ===\n")

	for i := 0; i < 3; i++ {
		go func(id int) {
			fmt.Printf("Goroutine %d: 开始执行\n", id)
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("Goroutine %d: 执行完成\n", id)
		}(i)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println()
}

// 演示2：Goroutine 的调度（观察执行顺序）
func demo2_Scheduling() {
	fmt.Println("=== 演示2：Goroutine 调度顺序（可能乱序）===\n")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d 执行\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println()
}

// 演示3：Goroutine 数量统计
func demo3_GoroutineCount() {
	fmt.Println("=== 演示3：Goroutine 数量统计 ===\n")

	fmt.Printf("初始 Goroutine 数: %d\n", runtime.NumGoroutine())

	// 创建 10 个 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()
	}

	fmt.Printf("创建 10 个 Goroutine 后: %d\n", runtime.NumGoroutine())
	wg.Wait()
	fmt.Printf("所有 Goroutine 完成后: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// 演示4：GOMAXPROCS 的影响
func demo4_GOMAXPROCS() {
	fmt.Println("=== 演示4：GOMAXPROCS（逻辑处理器数量）===\n")

	old := runtime.GOMAXPROCS(0)
	fmt.Printf("当前 GOMAXPROCS: %d (CPU 核心数)\n", old)
	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())

	// 设置 GOMAXPROCS = 2
	runtime.GOMAXPROCS(2)
	fmt.Printf("设置后 GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	// 恢复原值
	runtime.GOMAXPROCS(old)
	fmt.Println()
}

// 演示5：Goroutine 阻塞和唤醒（Channel）
func demo5_Blocking() {
	fmt.Println("=== 演示5：Goroutine 阻塞和唤醒（Channel）===\n")

	ch := make(chan int)

	// 发送方 goroutine（会阻塞直到有接收者）
	go func() {
		fmt.Println("发送方: 准备发送数据...")
		ch <- 42
		fmt.Println("发送方: 数据已发送")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("当前 Goroutine 数（发送方阻塞中）: %d\n", runtime.NumGoroutine())

	// 接收数据，唤醒发送方
	fmt.Println("接收方: 准备接收数据...")
	value := <-ch
	fmt.Printf("接收方: 收到数据 %d\n", value)

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("当前 Goroutine 数（发送方已唤醒）: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// 演示6：大量 Goroutine 的创建（展示轻量级特性）
func demo6_ManyGoroutines() {
	fmt.Println("=== 演示6：创建大量 Goroutine（展示轻量级）===\n")

	const count = 10000
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 简单计算
			_ = id * id
		}(i)
	}
	createCost := time.Since(start)

	fmt.Printf("创建 %d 个 Goroutine 耗时: %v\n", count, createCost)
	fmt.Printf("平均每个 Goroutine 创建耗时: %v\n", createCost/time.Duration(count))

	start = time.Now()
	wg.Wait()
	waitCost := time.Since(start)

	fmt.Printf("等待所有 Goroutine 完成耗时: %v\n", waitCost)
	fmt.Printf("当前 Goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// 演示7：Worker Pool（控制并发数）
func demo7_WorkerPool() {
	fmt.Println("=== 演示7：Worker Pool（控制并发数）===\n")

	const workerCount = 3
	const taskCount = 10

	taskChan := make(chan int, taskCount)
	resultChan := make(chan int, taskCount)

	// 启动固定数量的 worker
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskChan {
				fmt.Printf("Worker %d: 处理任务 %d\n", workerID, task)
				time.Sleep(50 * time.Millisecond) // 模拟工作
				resultChan <- task * 2
			}
		}(i)
	}

	// 发任务
	go func() {
		for i := 0; i < taskCount; i++ {
			taskChan <- i
		}
		close(taskChan)
	}()

	// 等待完成并关闭结果 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收结果
	fmt.Printf("Worker Pool 并发数: %d，任务数: %d\n", workerCount, taskCount)
	for result := range resultChan {
		fmt.Printf("收到结果: %d\n", result)
	}
	fmt.Println()
}

func main() {
	fmt.Println("Goroutine 工作原理演示\n")
	fmt.Println("=" + string(make([]byte, 50)) + "\n")

	demo1_BasicGoroutine()
	demo2_Scheduling()
	demo3_GoroutineCount()
	demo4_GOMAXPROCS()
	demo5_Blocking()
	demo6_ManyGoroutines()
	demo7_WorkerPool()

	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("\n演示完成！")
}
