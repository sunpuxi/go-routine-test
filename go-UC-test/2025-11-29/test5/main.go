package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ============================================
// 练习3（困难）：带重试、限流、错误处理的并发请求聚合
// 场景：真实生产环境，需要处理失败重试、控制并发数、收集所有结果
// 要求：
// 1) 使用 worker pool 控制并发数（例如最多3个并发请求）
// 2) 失败自动重试（最多重试2次）
// 3) 所有请求完成后，统计成功/失败数量
// ============================================

type RequestTask struct {
	URL     string
	Retries int // 剩余重试次数
}

type RequestResult struct {
	URL      string
	Content  string
	Error    error
	Attempts int // 实际尝试次数
	Cost     time.Duration
}

// 带重试的单个请求
func fetchWithRetry(ctx context.Context, client *http.Client, url string, maxRetries int) RequestResult {
	start := time.Now()
	var lastErr error
	attempts := 0

	for i := 0; i <= maxRetries; i++ {
		attempts++
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// TODO: 如果是超时错误，等待一小段时间再重试
			// 提示：可以用 time.Sleep(500 * time.Millisecond)
			if i < maxRetries {
				time.Sleep(500 * time.Millisecond)
			}
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// 成功
		return RequestResult{
			URL:      url,
			Content:  string(body)[:min(50, len(body))],
			Attempts: attempts,
			Cost:     time.Since(start),
		}
	}

	return RequestResult{
		URL:      url,
		Error:    lastErr,
		Attempts: attempts,
		Cost:     time.Since(start),
	}
}

// Worker pool 模式的并发请求（带限流）
func concurrentFetchWithWorkerPool(urls []string, maxWorkers int, maxRetries int, timeoutPerRequest time.Duration) []RequestResult {
	// 任务 channel
	taskChan := make(chan RequestTask, len(urls))
	// 结果 channel
	resultChan := make(chan RequestResult, len(urls))

	client := &http.Client{Timeout: 10 * time.Second}

	// 启动 worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// TODO: 从 taskChan 取任务，处理完后把结果发到 resultChan
			// 提示：for task := range taskChan { ... }
			for task := range taskChan {
				ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
				result := fetchWithRetry(ctx, client, task.URL, maxRetries)
				cancel()
				resultChan <- result
			}
		}()
	}

	// 主 goroutine：发任务（因为 taskChan 有缓冲，不会阻塞）
	for _, url := range urls {
		taskChan <- RequestTask{URL: url, Retries: maxRetries}
	}
	close(taskChan) // 关闭 taskChan，通知 worker 没有更多任务了

	// 启动一个 goroutine 等待所有 worker 完成，然后关闭 resultChan
	// 为什么需要这个 goroutine？
	// 因为主 goroutine 需要从 resultChan 收结果（下面的 for range），
	// 如果主 goroutine 直接 wg.Wait()，会阻塞，无法收结果
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 主 goroutine：收集结果（会阻塞直到 resultChan 关闭）
	results := make([]RequestResult, 0, len(urls))
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/2",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
	}

	maxWorkers := 3 // 最多3个并发
	maxRetries := 2 // 最多重试2次
	timeout := 3 * time.Second

	fmt.Printf("=== 练习3：带重试、限流的并发请求 ===\n")
	fmt.Printf("配置: workers=%d, maxRetries=%d, timeout=%v\n\n", maxWorkers, maxRetries, timeout)

	start := time.Now()
	results := concurrentFetchWithWorkerPool(urls, maxWorkers, maxRetries, timeout)
	totalCost := time.Since(start)

	successCount := 0
	failCount := 0
	totalAttempts := 0

	fmt.Println("详细结果：")
	for i, r := range results {
		if r.Error != nil {
			fmt.Printf("  [%d] %s: ❌ 失败 (尝试%d次, 耗时%v) - %v\n",
				i+1, r.URL, r.Attempts, r.Cost, r.Error)
			failCount++
		} else {
			fmt.Printf("  [%d] %s: ✅ 成功 (尝试%d次, 耗时%v)\n",
				i+1, r.URL, r.Attempts, r.Cost)
			successCount++
		}
		totalAttempts += r.Attempts
	}

	fmt.Printf("\n统计:\n")
	fmt.Printf("  总耗时: %v\n", totalCost)
	fmt.Printf("  成功: %d, 失败: %d\n", successCount, failCount)
	fmt.Printf("  总尝试次数: %d (平均每个请求 %.1f 次)\n",
		totalAttempts, float64(totalAttempts)/float64(len(urls)))
}
