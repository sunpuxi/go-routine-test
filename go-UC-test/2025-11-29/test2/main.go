package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	arraySize = 1_000_000
	// 每个元素要做的"重计算"次数（越大越慢，方便看出并发优势）
	heavyComputeRounds = 100
)

// 生成测试数组：1..arraySize
func generateArray() []int {
	arr := make([]int, arraySize)
	for i := 0; i < arraySize; i++ {
		arr[i] = i + 1
	}
	return arr
}

// 对一个元素做"假装很重"的计算
// 例如：循环做若干次平方/开方，最后返回一个结果值
func heavyCompute(x int) int {
	result := x
	// 循环做 heavyComputeRounds 次"重计算"
	for i := 0; i < heavyComputeRounds; i++ {
		result = result*result%1000 + 1
	}
	return result
}

// 串行版本：遍历数组，对每个元素做 heavyCompute，累加结果
func serialCompute(arr []int) int64 {
	var sum int64
	for _, item := range arr {
		sum += int64(heavyCompute(item))
	}
	return sum
}

// 并发版本：使用 worker pool 模式
// 思路：多个 worker goroutine 从 channel 里取任务（数组元素），计算后累加到局部 sum，最后汇总
func concurrentCompute(arr []int, workerCount int) int64 {
	if workerCount <= 0 {
		workerCount = 1
	}

	// 任务 channel：把数组元素发进去
	taskChan := make(chan int, 100) // 带缓冲，避免阻塞

	var wg sync.WaitGroup
	// 每个 worker 的局部 sum
	workerSums := make([]int64, workerCount)

	// 先启动所有 worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		workerIndex := i

		go func() {
			defer wg.Done()

			var localSum int64
			// 从 taskChan 里取任务，当 channel 关闭且取完所有任务后，退出循环
			for item := range taskChan {
				localSum += int64(heavyCompute(item))
			}
			// 把 localSum 写入 workerSums[workerIndex]
			workerSums[workerIndex] = localSum
		}()
	}

	// 主 goroutine：把数组元素发到 taskChan
	for _, item := range arr {
		taskChan <- item
	}
	// 发完后，关闭 channel（通知 worker 没有更多任务了）
	close(taskChan)

	// 等待所有 worker 完成
	wg.Wait()

	// 汇总所有 worker 的局部 sum
	var total int64
	for _, s := range workerSums {
		total += s
	}
	return total
}

func main() {
	arr := generateArray()
	fmt.Printf("数组长度: %d, 每个元素计算 %d 轮\n\n", len(arr), heavyComputeRounds)

	// 串行版本
	start := time.Now()
	serialResult := serialCompute(arr)
	serialCost := time.Since(start)

	// 并发版本（4 个 worker）
	start = time.Now()
	concurrentResult := concurrentCompute(arr, 4)
	concurrentCost := time.Since(start)

	fmt.Printf("串行结果: %d, 耗时: %v\n", serialResult, serialCost)
	fmt.Printf("并发结果: %d, 耗时: %v\n", concurrentResult, concurrentCost)

	if serialResult != concurrentResult {
		fmt.Println("❌ 结果不一致！检查并发逻辑")
	} else {
		fmt.Println("✅ 结果一致")
	}

	// 性能对比
	if concurrentCost > 0 {
		speedup := float64(serialCost) / float64(concurrentCost)
		fmt.Printf("加速比: %.2fx\n", speedup)
	}
}
