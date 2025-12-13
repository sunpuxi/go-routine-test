package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// 练习3（困难）：sync.Map 并发安全映射和 sync.Cond 条件变量
// 场景：高并发场景下的 map 操作，以及复杂的线程间协调需求
// 要求：
// 1）使用 sync.Map 实现一个线程安全的缓存，支持 Store、Load、Delete、Range 操作
// 2）对比 sync.Map 和 "map + Mutex" 在不同读写比例下的性能
// 3）使用 sync.Cond 实现一个生产者-消费者模型，当队列为空时消费者等待，队列满时生产者等待
// 4）理解 Cond 的使用模式：Wait() 前必须持有锁，Signal()/Broadcast() 唤醒等待的 goroutine
// 5）实现一个任务队列，支持阻塞式的 Put 和 Take 操作
// ============================================

// ============ 示例1：sync.Map 实现线程安全缓存 ============

// SafeCache 基于 sync.Map 的线程安全缓存
type SafeCache struct {
	m sync.Map
}

// NewSafeCache 创建新的缓存
func NewSafeCache() *SafeCache {
	return &SafeCache{}
}

// Set 存储键值对
func (c *SafeCache) Set(key, value interface{}) {
	c.m.Store(key, value)
}

// Get 获取值
func (c *SafeCache) Get(key interface{}) (interface{}, bool) {
	return c.m.Load(key)
}

// Delete 删除键值对
func (c *SafeCache) Delete(key interface{}) {
	c.m.Delete(key)
}

// Size 获取缓存大小（需要遍历，性能较差）
func (c *SafeCache) Size() int {
	count := 0
	c.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Range 遍历所有键值对
func (c *SafeCache) Range(fn func(key, value interface{}) bool) {
	c.m.Range(fn)
}

func testSyncMap() {
	fmt.Println("\n=== 测试：sync.Map 线程安全缓存 ===")
	cache := NewSafeCache()

	// 并发写入
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set(fmt.Sprintf("key%d", id), fmt.Sprintf("value%d", id))
		}(i)
	}
	wg.Wait()

	fmt.Printf("缓存大小: %d\n", cache.Size())

	// 并发读取
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			value, ok := cache.Get(fmt.Sprintf("key%d", id))
			if ok {
				fmt.Printf("读取 key%d: %v\n", id, value)
			}
		}(i)
	}
	wg.Wait()
}

// ============ 示例2：性能对比 sync.Map vs map+Mutex ============

// MutexMap 基于 Mutex 的线程安全 map
type MutexMap struct {
	mu sync.Mutex
	m  map[string]interface{}
}

func NewMutexMap() *MutexMap {
	return &MutexMap{
		m: make(map[string]interface{}),
	}
}

func (m *MutexMap) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

func (m *MutexMap) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.m[key]
	return value, ok
}

func benchmarkSyncMap(writeRatio float64, operations int) time.Duration {
	cache := NewSafeCache()
	start := time.Now()

	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/goroutines; j++ {
				key := fmt.Sprintf("key%d-%d", id, j)
				if float64(j%100)/100.0 < writeRatio {
					cache.Set(key, "value")
				} else {
					cache.Get(key)
				}
			}
		}(i)
	}
	wg.Wait()

	return time.Since(start)
}

func benchmarkMutexMap(writeRatio float64, operations int) time.Duration {
	mm := NewMutexMap()
	start := time.Now()

	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/goroutines; j++ {
				key := fmt.Sprintf("key%d-%d", id, j)
				if float64(j%100)/100.0 < writeRatio {
					mm.Set(key, "value")
				} else {
					mm.Get(key)
				}
			}
		}(i)
	}
	wg.Wait()

	return time.Since(start)
}

func testPerformance() {
	fmt.Println("\n=== 性能对比：sync.Map vs map+Mutex ===")
	operations := 10000

	// 多读少写场景
	fmt.Println("\n多读少写场景（写:读 = 1:9）:")
	duration1 := benchmarkSyncMap(0.1, operations)
	duration2 := benchmarkMutexMap(0.1, operations)
	fmt.Printf("sync.Map: %v\n", duration1)
	fmt.Printf("map+Mutex: %v\n", duration2)

	// 读写均衡场景
	fmt.Println("\n读写均衡场景（写:读 = 1:1）:")
	duration3 := benchmarkSyncMap(0.5, operations)
	duration4 := benchmarkMutexMap(0.5, operations)
	fmt.Printf("sync.Map: %v\n", duration3)
	fmt.Printf("map+Mutex: %v\n", duration4)

	// 多写少读场景
	fmt.Println("\n多写少读场景（写:读 = 9:1）:")
	duration5 := benchmarkSyncMap(0.9, operations)
	duration6 := benchmarkMutexMap(0.9, operations)
	fmt.Printf("sync.Map: %v\n", duration5)
	fmt.Printf("map+Mutex: %v\n", duration6)
}

// ============ 示例3：使用 sync.Cond 实现阻塞队列 ============

// BlockingQueue 基于 sync.Cond 的阻塞队列
type BlockingQueue struct {
	items    []interface{}
	mu       sync.Mutex
	notEmpty *sync.Cond // 队列非空的条件
	notFull  *sync.Cond // 队列未满的条件
	maxSize  int
}

// NewBlockingQueue 创建阻塞队列
func NewBlockingQueue(maxSize int) *BlockingQueue {
	bq := &BlockingQueue{
		items:   make([]interface{}, 0),
		maxSize: maxSize,
	}
	bq.notEmpty = sync.NewCond(&bq.mu)
	bq.notFull = sync.NewCond(&bq.mu)
	return bq
}

// Put 阻塞式添加元素
func (bq *BlockingQueue) Put(item interface{}) {
	bq.mu.Lock()
	defer bq.mu.Unlock()

	// 如果队列满了，等待
	for len(bq.items) >= bq.maxSize {
		bq.notFull.Wait() // 等待队列未满信号
	}

	bq.items = append(bq.items, item)
	bq.notEmpty.Signal() // 唤醒等待的消费者
}

// Take 阻塞式取出元素
func (bq *BlockingQueue) Take() interface{} {
	bq.mu.Lock()
	defer bq.mu.Unlock()

	// 如果队列为空，等待
	for len(bq.items) == 0 {
		bq.notEmpty.Wait() // 等待队列非空信号
	}

	item := bq.items[0]
	bq.items = bq.items[1:]
	bq.notFull.Signal() // 唤醒等待的生产者
	return item
}

// Size 获取队列大小
func (bq *BlockingQueue) Size() int {
	bq.mu.Lock()
	defer bq.mu.Unlock()
	return len(bq.items)
}

func testBlockingQueue() {
	fmt.Println("\n=== 测试：使用 sync.Cond 实现阻塞队列 ===")
	queue := NewBlockingQueue(5) // 最大容量为 5

	// 生产者
	var wg sync.WaitGroup
	producerCount := 3
	itemsPerProducer := 10

	for i := 0; i < producerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerProducer; j++ {
				item := fmt.Sprintf("生产者%d-项目%d", id, j)
				queue.Put(item)
				fmt.Printf("生产: %s (队列大小: %d)\n", item, queue.Size())
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	// 消费者
	consumerCount := 2
	for i := 0; i < consumerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < (producerCount*itemsPerProducer)/consumerCount; j++ {
				item := queue.Take()
				fmt.Printf("消费[%d]: %v (队列大小: %d)\n", id, item, queue.Size())
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("所有生产者和消费者完成")
}

// ============ 示例4：使用 Cond 实现任务队列 ============

// TaskQueue 任务队列
type TaskQueue struct {
	tasks   []func()
	mu      sync.Mutex
	hasTask *sync.Cond
	closed  bool
}

// NewTaskQueue 创建任务队列
func NewTaskQueue() *TaskQueue {
	tq := &TaskQueue{
		tasks: make([]func(), 0),
	}
	tq.hasTask = sync.NewCond(&tq.mu)
	return tq
}

// AddTask 添加任务
func (tq *TaskQueue) AddTask(task func()) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.closed {
		return
	}

	tq.tasks = append(tq.tasks, task)
	tq.hasTask.Signal() // 唤醒一个等待的 worker
}

// TakeTask 获取任务（阻塞）
func (tq *TaskQueue) TakeTask() (func(), bool) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// 如果队列为空且未关闭，等待
	for len(tq.tasks) == 0 && !tq.closed {
		tq.hasTask.Wait()
	}

	// 如果队列为空且已关闭，返回 nil
	if len(tq.tasks) == 0 {
		return nil, false
	}

	task := tq.tasks[0]
	tq.tasks = tq.tasks[1:]
	return task, true
}

// Close 关闭队列
func (tq *TaskQueue) Close() {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.closed = true
	tq.hasTask.Broadcast() // 唤醒所有等待的 worker
}

func testTaskQueue() {
	fmt.Println("\n=== 测试：任务队列 ===")
	queue := NewTaskQueue()

	// 启动 worker
	workerCount := 3
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				task, ok := queue.TakeTask()
				if !ok {
					fmt.Printf("Worker %d 退出\n", id)
					return
				}
				fmt.Printf("Worker %d 执行任务\n", id)
				task()
			}
		}(i)
	}

	// 添加任务
	for i := 0; i < 10; i++ {
		taskID := i
		queue.AddTask(func() {
			fmt.Printf("执行任务 %d\n", taskID)
			time.Sleep(100 * time.Millisecond)
		})
	}

	time.Sleep(2 * time.Second) // 等待任务完成
	queue.Close()               // 关闭队列
	wg.Wait()
	fmt.Println("所有 worker 已退出")
}

func main() {
	fmt.Println("【专题5 - 练习3：sync.Map 并发安全映射和 sync.Cond 条件变量】")

	// 测试 sync.Map
	testSyncMap()

	// 性能对比
	testPerformance()

	// 测试阻塞队列
	testBlockingQueue()

	// 测试任务队列
	testTaskQueue()

	fmt.Println("\n=== 练习提示 ===")
	fmt.Println("【sync.Map】")
	fmt.Println("1. sync.Map 适用于多读少写、键值对读写分离的场景")
	fmt.Println("2. 使用 Store、Load、Delete、Range 进行操作")
	fmt.Println("3. 在多读少写场景下性能优于 map+Mutex")
	fmt.Println("4. 不适合需要 Range 遍历所有键的场景（性能较差）")
	fmt.Println("\n【sync.Cond】")
	fmt.Println("1. Cond 需要配合 Mutex 使用，Wait() 前必须持有锁")
	fmt.Println("2. Signal() 唤醒一个等待的 goroutine，Broadcast() 唤醒所有")
	fmt.Println("3. 适用于复杂的线程间协调，如生产者-消费者、条件等待等")
	fmt.Println("4. 典型的等待模式：for condition { cond.Wait() }")
}
