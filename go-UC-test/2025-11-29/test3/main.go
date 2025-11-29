package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ============================================
// 练习1（简单）：并发发起多个 HTTP 请求
// 场景：需要从多个 URL 获取数据，串行太慢，用并发加速
// ============================================

// 串行版本：依次请求每个 URL
func serialFetch(urls []string) []string {
	results := make([]string, len(urls))
	client := &http.Client{Timeout: 5 * time.Second}

	for i, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			results[i] = fmt.Sprintf("ERROR: %v", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			results[i] = fmt.Sprintf("ERROR: %v", err)
			continue
		}
		results[i] = string(body)[:min(50, len(body))] // 只取前50字符
	}

	return results
}

// 并发版本：用 goroutine + WaitGroup 并发请求
func concurrentFetch(urls []string) []string {
	res := make([]string, len(urls))
	var wg sync.WaitGroup

	for i, url := range urls {
		// 并发任务加一
		wg.Add(1)

		go func(idx int, u string) {
			// 最后标记当前的并发任务结束
			defer wg.Done()
			rsp, err := http.Get(u)
			if err != nil {
				res[idx] = fmt.Sprintf("Error is %s", err.Error())
			}
			defer rsp.Body.Close()

			// 获取相应结果中的body
			body, err := io.ReadAll(rsp.Body)
			if err != nil {
				res[idx] = fmt.Sprintf("Error is %s", err.Error())
			}

			// 获取前50个字符
			res[idx] = string(body)[:min(50, len(body))]
		}(i, url)
	}

	// 等待所有的 goroutine 执行结束
	wg.Wait()

	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// 测试用的 URL（可以用 httpbin.org 这种测试服务）
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
	}

	fmt.Println("=== 练习1：并发 HTTP 请求 ===\n")

	// 串行版本
	start := time.Now()
	serialResults := serialFetch(urls)
	serialCost := time.Since(start)
	fmt.Printf("串行耗时: %v\n", serialCost)

	// 并发版本
	start = time.Now()
	concurrentResults := concurrentFetch(urls)
	concurrentCost := time.Since(start)
	fmt.Printf("并发耗时: %v\n", concurrentCost)

	if concurrentCost > 0 {
		speedup := float64(serialCost) / float64(concurrentCost)
		fmt.Printf("加速比: %.2fx\n", speedup)
	}

	fmt.Println("结果对比（前3个）：")
	for i := 0; i < min(3, len(urls)); i++ {
		fmt.Printf("URL %d: 串行=%s, 并发=%s\n", i+1,
			serialResults[i][:min(30, len(serialResults[i]))],
			concurrentResults[i][:min(30, len(concurrentResults[i]))])
	}
}
