package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// 练习2（中等）：使用 RWMutex 优化读写分离
// 场景：多读少写的场景，使用读写锁提升性能
// 要求：使用 sync.RWMutex 实现线程安全的缓存
// ============================================

// Cache 线程安全的缓存（使用读写锁）
type Cache struct {
	data map[string]string
	mu   sync.RWMutex // 读写锁
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

// Get 读取数据（使用读锁，允许多个 goroutine 同时读）
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

// Set 写入数据（使用写锁，独占访问）
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// 性能对比：使用普通 Mutex 的版本
type CacheWithMutex struct {
	data map[string]string
	mu   sync.Mutex // 普通互斥锁
}

func NewCacheWithMutex() *CacheWithMutex {
	return &CacheWithMutex{
		data: make(map[string]string),
	}
}

func (c *CacheWithMutex) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.data[key]
	return value, ok
}

func (c *CacheWithMutex) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func main() {
	// 验证读多写少的场景
	// 读操作的goroutine 的数量
	readCnt := 3000
	// 写操作对应的goroutine 的数量
	writeCnt := 10
	// 每个 goroutine 的操作次数
	operationCnt := 1000

	cache1 := NewCache()
	cache1.Set("Key1", "Value1")
	cache1.Set("Key2", "Value2")

	// read
	var wg1 sync.WaitGroup
	timeRead := time.Now()
	for i := 0; i < readCnt; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for j := 0; j < operationCnt; j++ {
				cache1.Get("Key1")
				cache1.Get("Key2")
			}
		}()
	}

	// write
	for i := 0; i < writeCnt; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for j := 0; j < operationCnt; j++ {
				cache1.Set("Key1", fmt.Sprintf("Value1-%d", j))
				cache1.Set("Key2", fmt.Sprintf("Value2-%d", j))
			}
		}()
	}

	wg1.Wait()
	fmt.Printf("write and read cost time is %v\n", time.Since(timeRead))

	// 普通的互斥锁
	cache2 := NewCacheWithMutex()
	cache2.Set("Key1", "Value1")
	cache2.Set("Key2", "Value2")

	// read
	var wg3 sync.WaitGroup
	timeRea2 := time.Now()
	for i := 0; i < readCnt; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			for j := 0; j < operationCnt; j++ {
				cache2.Get("Key1")
				cache2.Get("Key2")
			}
		}()
	}

	// write
	for i := 0; i < writeCnt; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			for j := 0; j < operationCnt; j++ {
				cache2.Set("Key1", fmt.Sprintf("Value1-%d", j))
				cache2.Set("Key2", fmt.Sprintf("Value2-%d", j))
			}
		}()
	}

	wg3.Wait()
	fmt.Printf("write and read cost time is %v\n", time.Since(timeRea2))
}
