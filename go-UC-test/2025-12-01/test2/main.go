package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// 练习2（中等）：扇入模式（Fan-In）
// 场景：多个 channel 的数据需要合并到一个 channel
// 要求：实现扇入模式，将多个输入 channel 合并为一个输出 channel
// ============================================

// fanIn 扇入函数：将多个输入 channel 合并为一个输出 channel
func fanIn(inputs ...<-chan int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup

	// 并发处理每个输入channel
	for _, in := range inputs {
		wg.Add(1)
		go func(ch <-chan int) {
			defer wg.Done()
			for v := range ch {
				output <- v
			}
		}(in)
	}

	// 等待所有goroutine完成，然后关闭输出channel
	go func() {
		wg.Wait()
		close(output)
	}()

	// 提前返回输出channel，让调用者可以立即使用
	return output
}

// generateNumbers 生成数字序列
func generateNumbers(name string, start, end int, delay time.Duration) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := start; i <= end; i++ {
			ch <- i
			time.Sleep(delay)
		}
		fmt.Printf("%s 完成\n", name)
	}()
	return ch
}

func main() {
	fmt.Println("=== 练习2：扇入模式（Fan-In）===\n")

	// 创建 3 个输入 channel
	ch1 := generateNumbers("Channel 1", 1, 5, 100*time.Millisecond)
	ch2 := generateNumbers("Channel 2", 6, 10, 150*time.Millisecond)
	ch3 := generateNumbers("Channel 3", 11, 16, 200*time.Millisecond)

	// 扇入：合并 3 个 channel
	merged := fanIn(ch1, ch2, ch3)

	// 从合并后的 channel 接收数据
	fmt.Println("接收合并后的数据：")
	for value := range merged {
		fmt.Printf("收到: %d\n", value)
	}

	fmt.Println("所有数据接收完成")
}
