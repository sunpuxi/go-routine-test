package main

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================
// 练习2（中等）：sync.Pool 实现对象池复用
// 场景：频繁创建和销毁对象会带来 GC 压力，使用对象池可以减少内存分配
// 要求：
// 1）实现一个基于 sync.Pool 的对象池，用于复用临时对象（如 buffer）
// 2）对比使用 Pool 和不使用 Pool 的性能差异（内存分配次数、GC 压力）
// 3）理解 Pool 的工作原理：Get() 从池中获取，Put() 归还对象
// 4）注意：Pool 中的对象随时可能被 GC 回收，不能假设对象一直存在
// ============================================

// ============ 示例1：不使用 Pool 的版本 ============

// processWithoutPool 不使用对象池的处理函数
func processWithoutPool(data []byte) []byte {
	// 每次都创建新的 buffer
	buf := make([]byte, 0, 1024)
	buf = append(buf, []byte("前缀: ")...)
	buf = append(buf, data...)
	buf = append(buf, []byte(" :后缀")...)
	return buf
}

// ============ 示例2：使用 Pool 的版本 ============

// bufferPool 全局的 buffer 对象池
var bufferPool = sync.Pool{
	New: func() interface{} {
		// 当池中没有对象时，会调用 New 函数创建新对象
		return &bytes.Buffer{}
	},
}

// processWithPool 使用对象池的处理函数
func processWithPool(data []byte) []byte {
	// 从池中获取 buffer
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf) // 处理完后归还到池中
	
	// 重置 buffer（重要：复用前要清空）
	buf.Reset()
	buf.WriteString("前缀: ")
	buf.Write(data)
	buf.WriteString(" :后缀")
	
	// 返回副本，因为 buf 会被归还到池中
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// ============ 示例3：更通用的对象池封装 ============

// StringBufferPool 字符串缓冲区池
type StringBufferPool struct {
	pool sync.Pool
}

// NewStringBufferPool 创建字符串缓冲区池
func NewStringBufferPool() *StringBufferPool {
	return &StringBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get 从池中获取 buffer
func (p *StringBufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put 归还 buffer 到池中
func (p *StringBufferPool) Put(buf *bytes.Buffer) {
	buf.Reset() // 清空内容
	p.pool.Put(buf)
}

// ============ 性能测试函数 ============

// runTest 运行性能测试
func runTest(name string, fn func()) {
	// 触发一次 GC，确保测试环境一致
	runtime.GC()
	runtime.GC()
	
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	start := time.Now()
	fn()
	duration := time.Since(start)
	
	runtime.ReadMemStats(&m2)
	
	fmt.Printf("\n=== %s ===\n", name)
	fmt.Printf("耗时: %v\n", duration)
	fmt.Printf("内存分配次数: %d\n", m2.Mallocs-m1.Mallocs)
	fmt.Printf("内存分配大小: %d KB\n", (m2.TotalAlloc-m1.TotalAlloc)/1024)
	fmt.Printf("GC 次数: %d\n", m2.NumGC-m1.NumGC)
}

func testWithoutPool() {
	const iterations = 10000
	data := []byte("测试数据")
	
	runTest("不使用 Pool", func() {
		for i := 0; i < iterations; i++ {
			_ = processWithoutPool(data)
		}
	})
}

func testWithPool() {
	const iterations = 10000
	data := []byte("测试数据")
	
	runTest("使用 Pool", func() {
		for i := 0; i < iterations; i++ {
			_ = processWithPool(data)
		}
	})
}

// ============ 示例4：实际应用场景 - 字符串拼接 ============

var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// buildStringWithPool 使用 Pool 构建字符串
func buildStringWithPool(parts []string) string {
	buf := stringBuilderPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		stringBuilderPool.Put(buf)
	}()
	
	for _, part := range parts {
		buf.WriteString(part)
	}
	return buf.String()
}

// buildStringWithoutPool 不使用 Pool 构建字符串
func buildStringWithoutPool(parts []string) string {
	var buf bytes.Buffer
	for _, part := range parts {
		buf.WriteString(part)
	}
	return buf.String()
}

func testStringBuilding() {
	fmt.Println("\n=== 测试：字符串拼接场景 ===")
	parts := []string{"Hello", " ", "World", "!", " ", "This", " ", "is", " ", "a", " ", "test"}
	
	var wg sync.WaitGroup
	const goroutines = 100
	const iterationsPerGoroutine = 100
	
	// 测试不使用 Pool
	start1 := time.Now()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				_ = buildStringWithoutPool(parts)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("不使用 Pool: %v\n", time.Since(start1))
	
	// 测试使用 Pool
	start2 := time.Now()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				_ = buildStringWithPool(parts)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("使用 Pool: %v\n", time.Since(start2))
}

func main() {
	fmt.Println("【专题5 - 练习2：sync.Pool 实现对象池复用】")
	
	// 性能对比测试
	testWithoutPool()
	testWithPool()
	
	// 字符串拼接场景测试
	testStringBuilding()
	
	fmt.Println("\n=== 练习提示 ===")
	fmt.Println("1. sync.Pool 用于缓存临时对象，减少内存分配和 GC 压力")
	fmt.Println("2. Get() 从池中获取对象，如果池为空则调用 New 函数创建")
	fmt.Println("3. Put() 归还对象到池中，注意要先 Reset() 清空对象状态")
	fmt.Println("4. Pool 中的对象随时可能被 GC 回收，不能依赖对象的生命周期")
	fmt.Println("5. 适用于：频繁创建销毁的对象（如 buffer、slice、struct）")
	fmt.Println("6. 不适合：需要长期持有的对象，或者有状态的资源")
}

