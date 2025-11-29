package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================
// 练习2：带超时和错误处理的 Worker Pool
// ============================================

// Task 任务结构
type Task struct {
	ID      int           // 任务ID
	Timeout time.Duration // 该任务的超时时间
}

// Result 结果结构
type Result struct {
	TaskID int           // 任务ID
	Value  int           // 成功时的结果值
	Error  error         // 失败时的错误信息
	Cost   time.Duration // 实际耗时
}

// processTask 处理单个任务
// 如果任务ID是偶数，有30%概率失败
// 使用 context 控制超时
func processTask(ctx context.Context, task Task) Result {
	start := time.Now()

	// 模拟任务处理耗时
	// 实际处理时间 = 超时时间的 0.5-1.5 倍（可能超时，也可能正常完成）
	actualDuration := task.Timeout/2 + time.Duration(rand.Float64()*float64(task.Timeout))

	// 使用 select 同时监听：任务完成和超时
	done := make(chan bool)
	go func() {
		time.Sleep(actualDuration)
		done <- true
	}()

	select {
	case <-done:
		// 任务正常完成（在超时前完成）
		// 检查是否是偶数ID，如果是，30%概率失败
		if task.ID%2 == 0 {
			if rand.Float32() < 0.3 { // 30% 概率失败
				return Result{
					TaskID: task.ID,
					Error:  fmt.Errorf("模拟错误：任务 %d 处理失败", task.ID),
					Cost:   time.Since(start),
				}
			}
		}

		// 成功
		return Result{
			TaskID: task.ID,
			Value:  task.ID * task.ID,
			Cost:   time.Since(start),
		}
	case <-ctx.Done():
		// 超时或被取消
		return Result{
			TaskID: task.ID,
			Error:  context.DeadlineExceeded,
			Cost:   time.Since(start),
		}
	}
}

// WorkerPool 带超时和错误处理的 Worker Pool
func WorkerPool(tasks []Task, workerCount int) []Result {
	// 创建 channel
	taskChan := make(chan Task, len(tasks))
	resultChan := make(chan Result, len(tasks))

	var wg sync.WaitGroup

	// 启动 Worker
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for task := range taskChan {
				fmt.Printf("Worker %d: 开始处理任务 %d (超时: %v)\n",
					workerID, task.ID, task.Timeout)

				// 为每个任务创建独立的超时 context
				ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
				result := processTask(ctx, task)
				cancel()

				// 输出处理结果
				if result.Error != nil {
					if result.Error == context.DeadlineExceeded {
						fmt.Printf("Worker %d: 任务 %d 超时 (耗时: %v)\n",
							workerID, task.ID, result.Cost)
					} else {
						fmt.Printf("Worker %d: 任务 %d 失败: %v (耗时: %v)\n",
							workerID, task.ID, result.Error, result.Cost)
					}
				} else {
					fmt.Printf("Worker %d: 任务 %d 完成，结果 = %d (耗时: %v)\n",
						workerID, task.ID, result.Value, result.Cost)
				}

				resultChan <- result
			}
		}(i)
	}

	// 发送任务
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 等待所有 Worker 完成并关闭结果 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	results := make([]Result, 0, len(tasks))
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// generateTasks 生成任务列表
func generateTasks(count int) []Task {
	tasks := make([]Task, count)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		// 随机超时时间：50ms - 500ms
		timeout := time.Duration(50+rand.Intn(450)) * time.Millisecond
		tasks[i] = Task{
			ID:      i,
			Timeout: timeout,
		}
	}

	return tasks
}

// printStatistics 打印统计信息
func printStatistics(results []Result, totalTime time.Duration) {
	successCount := 0
	failCount := 0
	timeoutCount := 0
	totalCost := time.Duration(0)

	for _, r := range results {
		if r.Error != nil {
			// 判断是否是超时错误
			if r.Error == context.DeadlineExceeded {
				timeoutCount++
			} else {
				failCount++
			}
		} else {
			successCount++
		}
		totalCost += r.Cost
	}

	avgCost := totalCost / time.Duration(len(results))

	fmt.Println("\n=== 统计结果 ===")
	fmt.Printf("总任务数: %d\n", len(results))
	fmt.Printf("成功: %d\n", successCount)
	fmt.Printf("失败: %d\n", failCount)
	fmt.Printf("超时: %d\n", timeoutCount)
	fmt.Printf("总耗时: %v\n", totalTime)
	fmt.Printf("平均耗时: %v\n", avgCost)
}

func main() {
	const workerCount = 4
	const taskCount = 20

	fmt.Println("=== Worker Pool 启动 ===")
	fmt.Printf("Worker 数量: %d\n", workerCount)
	fmt.Printf("任务数量: %d\n\n", taskCount)

	// 生成任务
	tasks := generateTasks(taskCount)

	// 执行 Worker Pool
	start := time.Now()
	results := WorkerPool(tasks, workerCount)
	totalTime := time.Since(start)

	// 打印统计信息
	printStatistics(results, totalTime)
}
