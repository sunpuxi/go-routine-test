package main

import (
	"fmt"
	"time"
)

// ============================================
// 练习1（简单）：使用 select 实现超时控制
// 场景：从多个 channel 接收数据，需要设置超时
// 要求：使用 select 语句实现非阻塞接收和超时
// ============================================

// fetchData 模拟从 channel 获取数据（可能很慢）
func fetchData(ch chan string) {
	time.Sleep(2 * time.Second) // 模拟耗时操作
	ch <- "数据获取成功"
}

// fetchWithTimeout 带超时的数据获取
func fetchWithTimeout(timeout time.Duration) (string, error) {
	dataChan := make(chan string)

	// 单独开一个goroutine获取数据
	go fetchData(dataChan)

	select {
	case data := <-dataChan:
		return data, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("time out. timeout is %v", timeout)
	}
}

// 演示：select 的多路复用
func demoSelectMultiplex() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	// 启动两个 goroutine 分别发送数据
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "来自 ch1 的数据"
	}()
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "来自 ch2 的数据"
	}()

	select {
	case data := <-ch1:
		fmt.Printf("从ch1收到: %s\n", data)
	case data := <-ch2:
		fmt.Printf("从ch2收到: %s\n", data)
	}
}

func main() {
	fmt.Println("=== 练习1：select 实现超时控制 ===\n")

	// 测试1：正常情况（数据在超时前返回）
	fmt.Println("测试1：正常情况（超时时间 3 秒）")
	result, err := fetchWithTimeout(3 * time.Second)
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
	} else {
		fmt.Printf("✅ 成功: %s\n", result)
	}

	// 测试2：超时情况（数据在超时后返回）
	fmt.Println("\n测试2：超时情况（超时时间 1 秒）")
	result, err = fetchWithTimeout(1 * time.Second)
	if err != nil {
		fmt.Printf("✅ 正确超时: %v\n", err)
	} else {
		fmt.Printf("❌ 应该超时但返回了: %s\n", result)
	}

	// 测试3：select 多路复用
	fmt.Println("\n测试3：select 多路复用")
	demoSelectMultiplex()
}
