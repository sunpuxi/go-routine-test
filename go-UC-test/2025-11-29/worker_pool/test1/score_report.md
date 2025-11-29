# 练习1 评分报告

## 总分：90/100

---

## 详细评分

### 1. 正确实现 Worker Pool 模式（20/20）✅

**优点**：
- ✅ 正确创建了固定数量的 Worker（通过参数 `workerCnt` 配置）
- ✅ Worker 使用 `for range taskChan` 从任务 channel 取任务
- ✅ 每个 Worker 独立处理任务，互不干扰

**代码片段**：
```go
for i := 0; i < workerCnt; i++ {
    wg.Add(1)
    go func(workerId int) {
        defer wg.Done()
        for task := range taskChan {
            // 处理任务
        }
    }(i)
}
```

---

### 2. 正确使用 channel 和 goroutine（20/20）✅

**优点**：
- ✅ 正确创建了带缓冲的 channel（`make(chan int, taskCnt)`）
- ✅ 正确使用 goroutine 启动 Worker
- ✅ 正确使用闭包传递 `workerId`，避免变量捕获问题
- ✅ 正确使用 channel 发送和接收数据

**代码片段**：
```go
taskChan := make(chan int, taskCnt)
resChan := make(chan int, taskCnt)
// ... 正确使用
```

---

### 3. 正确关闭 channel，无死锁（20/20）✅

**优点**：
- ✅ 发送完所有任务后，正确关闭了 `taskChan`（第32行）
- ✅ 使用单独的 goroutine 等待所有 Worker 完成，然后关闭 `resChan`（第34-37行）
- ✅ 主 goroutine 使用 `for range resChan` 收集结果
- ✅ 没有死锁风险

**关键代码**：
```go
close(taskChan)  // ✅ 正确关闭

go func() {
    wg.Wait()
    close(resChan)  // ✅ 正确关闭
}()
```

---

### 4. 正确使用 WaitGroup 等待所有 Worker 完成（20/20）✅

**优点**：
- ✅ 正确使用 `wg.Add(1)` 增加计数器
- ✅ 正确使用 `defer wg.Done()` 减少计数器
- ✅ 正确使用单独的 goroutine 执行 `wg.Wait()`，避免阻塞主 goroutine

**代码片段**：
```go
wg.Add(1)
go func(workerId int) {
    defer wg.Done()
    // ...
}(i)
```

---

### 5. 代码结构清晰，注释合理（10/20）⚠️

**优点**：
- ✅ 代码结构清晰，函数命名合理（`DoFuncByWorkerPool`）
- ✅ 变量命名清晰（`taskChan`、`resChan`、`workerId`）

**问题**：
- ❌ **缺少注释**：没有函数注释、关键逻辑注释
- ❌ **缺少模拟耗时**：题目明确要求"每个任务处理需要 100ms"，但代码中没有 `time.Sleep(100 * time.Millisecond)`
- ⚠️ 输出格式与题目示例略有差异（但功能正确）

**改进建议**：
```go
// DoFuncByWorkerPool 使用 Worker Pool 模式并发处理任务
func DoFuncByWorkerPool(workerCnt int) {
    // ...
    for task := range taskChan {
        fmt.Printf("Worker %d: 开始处理任务 %d\n", workerId, task)
        time.Sleep(100 * time.Millisecond)  // ← 添加模拟耗时
        result := task * task
        fmt.Printf("Worker %d: 任务 %d 完成，结果 = %d\n", workerId, task, result)
        resChan <- result
    }
}
```

---

## 代码问题总结

### 必须修复的问题

1. **缺少模拟耗时**（-10分）
   - 题目要求：每个任务处理需要 100ms
   - 当前代码：没有 `time.Sleep`
   - 修复：在任务处理中添加 `time.Sleep(100 * time.Millisecond)`

### 建议改进

1. **添加注释**
   - 函数注释说明功能
   - 关键逻辑注释

2. **输出格式优化**
   - 可以添加"开始处理"和"完成"的提示
   - 主 goroutine 收集结果时可以添加"收到结果"前缀

---

## 改进后的代码示例

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	taskCnt = 10
)

// DoFuncByWorkerPool 使用 Worker Pool 模式并发处理任务
// workerCnt: Worker 数量
func DoFuncByWorkerPool(workerCnt int) {
	var wg sync.WaitGroup

	// 创建任务和结果 channel
	taskChan := make(chan int, taskCnt)
	resChan := make(chan int, taskCnt)

	// 启动 Worker
	for i := 0; i < workerCnt; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for task := range taskChan {
				fmt.Printf("Worker %d: 开始处理任务 %d\n", workerId, task)
				time.Sleep(100 * time.Millisecond) // 模拟耗时
				result := task * task
				fmt.Printf("Worker %d: 任务 %d 完成，结果 = %d\n", workerId, task, result)
				resChan <- result
			}
		}(i)
	}

	// 发送任务
	for i := range taskCnt {
		taskChan <- i
	}
	close(taskChan)

	// 等待所有 Worker 完成并关闭结果 channel
	go func() {
		wg.Wait()
		close(resChan)
	}()

	// 收集结果
	for item := range resChan {
		fmt.Printf("主 goroutine: 收到结果 %d\n", item)
	}
	fmt.Println("主 goroutine: 所有任务完成")
}

func main() {
	DoFuncByWorkerPool(3) // 使用 3 个 Worker（符合题目示例）
}
```

---

## 总结

**优点**：
- ✅ Worker Pool 模式实现正确
- ✅ Channel 和 Goroutine 使用正确
- ✅ 无死锁风险
- ✅ WaitGroup 使用正确

**需要改进**：
- ❌ 添加模拟耗时（100ms）
- ⚠️ 添加注释提升可读性

**总体评价**：代码实现正确，核心功能完整，但缺少题目要求的模拟耗时。建议添加 `time.Sleep` 后再提交。

