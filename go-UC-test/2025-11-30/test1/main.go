package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// 练习1（简单）：使用 Mutex 保护共享变量
// 场景：多个 goroutine 同时修改一个计数器
// 要求：使用 sync.Mutex 保证并发安全
// ============================================

// Counter 线程安全的计数器
type Counter struct {
	value int
	mu    sync.Mutex // 互斥锁
}

// Inc 增加计数（线程安全）
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Get 获取当前值（线程安全）
func (c *Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// 错误的实现：不使用锁（会有竞态条件）
type UnsafeCounter struct {
	value int
}

func (c *UnsafeCounter) Inc() {
	c.value++ // ❌ 并发不安全
}

func (c *UnsafeCounter) Get() int {
	return c.value // ❌ 并发不安全
}

func main() {
	// goroutine 的数量
	goroutineNum := 1000
	// 相加的次数
	taskCount := 1000
	// 线程安全的counter（Mutex 零值可用，不需要显式初始化）
	safeCounter := &Counter{}
	// 开启goroutine执行相加
	var wg sync.WaitGroup
	time1 := time.Now()
	for i := 0; i < goroutineNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < taskCount; j++ {
				safeCounter.Inc()
			}
		}()
	}

	wg.Wait()
	fmt.Printf("safeCounter.Get() = %d,cost time is %v\n", safeCounter.Get(), time.Since(time1))

	// 线程不安全的 Counter
	counter := &UnsafeCounter{value: 0}
	var wg2 sync.WaitGroup
	time2 := time.Now()
	for i := 0; i < goroutineNum; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done() // ✅ 必须调用 Done()，否则 wg2.Wait() 会永远阻塞
			for j := 0; j < taskCount; j++ {
				counter.Inc()
			}
		}()
	}

	wg2.Wait()
	fmt.Printf("counter.Get() = %d,cost time is %v\n", counter.Get(), time.Since(time2))
}
