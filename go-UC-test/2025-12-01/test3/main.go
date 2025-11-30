package main

import (
	"fmt"
	"time"
)

// ============================================
// 练习3（困难）：管道模式（Pipeline）
// 场景：数据处理流水线，多个阶段依次处理数据
// 要求：实现一个管道，数据依次经过多个处理阶段
// ============================================

// Stage 处理阶段函数类型
type Stage func(<-chan int) <-chan int

// square 阶段1：计算平方
func square(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			result := value * value
			fmt.Printf("  [平方阶段] %d -> %d\n", value, result)
			output <- result
			time.Sleep(50 * time.Millisecond) // 模拟处理耗时
		}
	}()
	return output
}

// add 阶段2：加 10
func add(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			result := value + 10
			fmt.Printf("  [加法阶段] %d -> %d\n", value, result)
			output <- result
			time.Sleep(50 * time.Millisecond) // 模拟处理耗时
		}
	}()
	return output
}

// multiply 阶段3：乘以 2
func multiply(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			result := value * 2
			fmt.Printf("  [乘法阶段] %d -> %d\n", value, result)
			output <- result
			time.Sleep(50 * time.Millisecond) // 模拟处理耗时
		}
	}()
	return output
}

// pipeline 构建管道
func pipeline(input <-chan int, stages ...Stage) <-chan int {
	current := input
	for _, stage := range stages {
		current = stage(current)
	}

	return current
}

// generateInput 生成输入数据 生产者，构建数据的过程异步执行
func generateInput(count int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= count; i++ {
			fmt.Printf("输入: %d\n", i)
			ch <- i
		}
	}()
	return ch
}

func main() {
	fmt.Println("=== 练习3：管道模式（Pipeline）===\n")
	fmt.Println("管道流程：输入 -> 平方 -> 加10 -> 乘以2 -> 输出\n")

	// 生成输入数据
	input := generateInput(5)

	// 构建管道：square -> add -> multiply
	output := pipeline(input, square, add, multiply)

	// 收集结果
	fmt.Println("\n最终结果：")
	results := make([]int, 0)
	for value := range output {
		results = append(results, value)
		fmt.Printf("✅ 最终输出: %d\n", value)
	}

	fmt.Printf("\n处理完成，共 %d 个结果\n", len(results))

	// 验证：对于输入 3
	// 平方: 3 * 3 = 9
	// 加10: 9 + 10 = 19
	// 乘以2: 19 * 2 = 38
	fmt.Println("\n验证（输入 3）：")
	fmt.Println("  3 -> 平方 -> 9 -> 加10 -> 19 -> 乘以2 -> 38")
}
