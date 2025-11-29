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
// 练习2（中等）：带超时控制的并发请求
// 场景：多个请求中，有些可能很慢或卡死，需要设置超时，超时的请求直接放弃
// 关键：使用 context.WithTimeout 控制单个请求的超时
// ============================================

type Result struct {
	URL     string
	Content string
	Error   error
	Cost    time.Duration
}

// 带超时的单个请求
func fetchWithTimeout(ctx context.Context, client *http.Client, url string) Result {
	start := time.Now()

	// TODO: 创建一个带超时的 request
	// 提示：http.NewRequestWithContext(ctx, "GET", url, nil)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{URL: url, Error: err, Cost: time.Since(start)}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err, Cost: time.Since(start)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{URL: url, Error: err, Cost: time.Since(start)}
	}

	return Result{
		URL:     url,
		Content: string(body)[:min(50, len(body))],
		Cost:    time.Since(start),
	}
}

// 并发请求，每个请求有独立超时
func concurrentFetchWithTimeout(urls []string, timeoutPerRequest time.Duration) []Result {
	results := make([]Result, len(urls))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 10 * time.Second} // 客户端总超时（兜底）

	for i, url := range urls {
		wg.Add(1)

		go func(index int, u string) {
			defer wg.Done()

			// TODO: 为每个请求创建独立的 context.WithTimeout
			// 提示：ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
			// defer cancel()
			// 然后调用 fetchWithTimeout(ctx, client, u)
			ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
			defer cancel()

			results[index] = fetchWithTimeout(ctx, client, u)
		}(i, url)
	}

	wg.Wait()
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
		"https://httpbin.org/delay/1", // 正常请求，1秒返回
		"https://httpbin.org/delay/2", // 正常请求，2秒返回
		"https://httpbin.org/delay/5", // 会超时（如果超时设为3秒）
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
	}

	timeout := 3 * time.Second
	fmt.Printf("=== 练习2：带超时控制的并发请求（超时=%v）===\n\n", timeout)

	start := time.Now()
	results := concurrentFetchWithTimeout(urls, timeout)
	totalCost := time.Since(start)

	fmt.Printf("总耗时: %v\n\n", totalCost)

	successCount := 0
	timeoutCount := 0
	for i, r := range results {
		if r.Error != nil {
			fmt.Printf("请求 %d [%s]: ❌ 错误 - %v (耗时: %v)\n", i+1, r.URL, r.Error, r.Cost)
			if r.Error == context.DeadlineExceeded {
				timeoutCount++
			}
		} else {
			fmt.Printf("请求 %d [%s]: ✅ 成功 (耗时: %v, 内容长度: %d)\n",
				i+1, r.URL, r.Cost, len(r.Content))
			successCount++
		}
	}

	fmt.Printf("\n统计: 成功=%d, 超时=%d, 总耗时=%v\n", successCount, timeoutCount, totalCost)
}
