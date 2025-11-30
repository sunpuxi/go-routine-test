package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// 练习3（困难）：原子操作 vs Mutex 性能对比
// 场景：简单的计数器操作，对比 atomic 和 mutex 的性能
// 要求：实现两种线程安全的计数器，对比性能
// ============================================

// CounterWithMutex 使用 Mutex 的计数器
type CounterWithMutex struct {
	value int64
	mu    sync.Mutex
}

func (c *CounterWithMutex) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *CounterWithMutex) Get() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// CounterWithAtomic 使用 Atomic 的计数器
type CounterWithAtomic struct {
	value int64 // 必须用 int64，atomic 只支持特定类型
}

func (c *CounterWithAtomic) Inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *CounterWithAtomic) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// CounterWithAtomicCAS 使用 CAS（Compare-And-Swap）的计数器（高级用法）
type CounterWithAtomicCAS struct {
	value int64
}

func (c *CounterWithAtomicCAS) Inc() {
	// TODO: 使用 atomic.CompareAndSwapInt64 实现自增
	// 提示：循环直到 CAS 成功
	for {
		old := atomic.LoadInt64(&c.value)
		new := old + 1
		if atomic.CompareAndSwapInt64(&c.value, old, new) {
			return
		}
	}
}

func (c *CounterWithAtomicCAS) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

func benchmarkCounter(name string, counter interface {
	Inc()
	Get() int64
}, goroutineCount, operationsPerGoroutine int) time.Duration {
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				counter.Inc()
			}
		}()
	}
	wg.Wait()
	cost := time.Since(start)

	expected := int64(goroutineCount * operationsPerGoroutine)
	actual := counter.Get()
	if actual != expected {
		fmt.Printf("  ❌ %s: 结果错误，期望 %d，实际 %d\n", name, expected, actual)
	} else {
		fmt.Printf("  ✅ %s: 结果正确，耗时 %v\n", name, cost)
	}

	return cost
}

func main() {
	const goroutineCount = 100
	const operationsPerGoroutine = 10000

	fmt.Println("=== 练习3：Atomic vs Mutex 性能对比 ===\n")
	fmt.Printf("配置: %d 个 goroutine，每个执行 %d 次操作\n\n", goroutineCount, operationsPerGoroutine)

	// 测试1：Mutex
	fmt.Println("测试1：使用 Mutex")
	counterMutex := &CounterWithMutex{}
	costMutex := benchmarkCounter("Mutex", counterMutex, goroutineCount, operationsPerGoroutine)

	// 测试2：Atomic
	fmt.Println("\n测试2：使用 Atomic")
	counterAtomic := &CounterWithAtomic{}
	costAtomic := benchmarkCounter("Atomic", counterAtomic, goroutineCount, operationsPerGoroutine)

	// 测试3：Atomic CAS（高级）
	fmt.Println("\n测试3：使用 Atomic CAS")
	counterCAS := &CounterWithAtomicCAS{}
	costCAS := benchmarkCounter("Atomic CAS", counterCAS, goroutineCount, operationsPerGoroutine)

	// 性能对比
	fmt.Println("\n=== 性能对比 ===")
	fmt.Printf("Mutex:      %v\n", costMutex)
	fmt.Printf("Atomic:     %v\n", costAtomic)
	fmt.Printf("Atomic CAS: %v\n", costCAS)

	if costAtomic < costMutex {
		speedup := float64(costMutex) / float64(costAtomic)
		fmt.Printf("\n✅ Atomic 比 Mutex 快 %.2fx\n", speedup)
		fmt.Println("原因：Atomic 是 CPU 级别的原子操作，无需系统调用")
	}

	if costCAS > costAtomic {
		fmt.Println("\n⚠️ Atomic CAS 可能比直接 Atomic 慢（因为循环重试）")
		fmt.Println("CAS 适用于复杂的原子操作，简单自增用 AddInt64 更高效")
	}
}
