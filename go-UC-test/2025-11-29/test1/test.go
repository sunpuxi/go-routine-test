package main

import (
	"fmt"
	"sync"
	"time"
)

// 单线程版本：计算 1..N 的平方和
func serialSumOfSquares(n int64) int64 {
	var sum int64
	for i := int64(1); i <= n; i++ {
		sum += i * i
	}
	return sum
}

// 并发版本：把 [1..N] 切成多个区间，每个 goroutine 负责一段
func concurrentSumOfSquares(n int64, workerCount int) int64 {
	if workerCount <= 0 {
		workerCount = 1
	}

	var wg sync.WaitGroup
	partSums := make([]int64, workerCount)

	// TODO: 计算每个 worker 处理的区间大小
	chunkSize := n / int64(workerCount)
	// 注意：最后一个 worker 可能需要吃掉剩余的数字

	for i := 0; i < workerCount; i++ {
		// wg + 1
		wg.Add(1)

		// TODO: 计算当前 worker 的 start / end（都是 int64）
		workerIndex := i
		var start, end int64

		// 示例：大致结构是这样，具体区间你来填
		// start = ...
		// end = ...
		// 左闭右开的区间
		start = int64(workerIndex) * chunkSize
		end = start + chunkSize - 1

		go func() {
			defer wg.Done()

			var localSum int64
			for j := start; j <= end; j++ {
				localSum += j * j
			}
			partSums[workerIndex] = localSum
		}()
	}

	// 等待全部的goroutine 执行完毕
	wg.Wait()

	// 统计这几个goroutine 的执行结果
	var total int64
	for _, v := range partSums {
		total += v
	}
	return total + n*n
}

func main() {
	const N int64 = 1_000_000
	const workers = 4

	// 串行
	start := time.Now()
	serial := serialSumOfSquares(N)
	serialCost := time.Since(start)

	// 并发
	start = time.Now()
	concurrent := concurrentSumOfSquares(N, workers)
	concurrentCost := time.Since(start)

	fmt.Println("serial    =", serial, "cost =", serialCost)
	fmt.Println("concurrent=", concurrent, "cost =", concurrentCost)

	if serial != concurrent {
		fmt.Println("结果不一致，说明并发实现有 bug，需要排查区间或汇总逻辑。")
	}
}
