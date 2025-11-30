# 练习2 评分报告

## 总分：75/100

---

## 详细评分

### 1. RWMutex 使用正确性（30/30）✅

**优点**：
- ✅ `Get()` 正确使用 `RLock()` 和 `RUnlock()`
- ✅ `Set()` 正确使用 `Lock()` 和 `Unlock()`
- ✅ 理解读锁和写锁的区别

**代码片段**：
```go
func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()        // ✅ 读锁
    defer c.mu.RUnlock()
    // ...
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()         // ✅ 写锁
    defer c.mu.Unlock()
    // ...
}
```

---

### 2. 并发安全实现（20/20）✅

**优点**：
- ✅ 所有对共享数据的访问都被正确保护
- ✅ 读操作使用读锁，写操作使用写锁
- ✅ 对比版本（Mutex）也正确实现

---

### 3. 测试逻辑设计（15/30）⚠️

**问题**：
- ❌ **关键问题**：读和写是**分开测试**的，而不是同时进行
- ❌ RWMutex 的优势在于"多读少写**同时进行**"的场景
- ❌ 如果读和写分开测试，RWMutex 和 Mutex 的性能差异不明显

**当前代码**：
```go
// 先测试读
wg1.Wait()  // 等待所有读完成
// 再测试写
wg2.Wait()  // 等待所有写完成
```

**应该改为**：
```go
// 同时启动读和写的 goroutine
// 这样才能体现出 RWMutex 的优势
```

**为什么重要？**
- RWMutex 的优势：多个读操作可以**同时进行**（不互相阻塞）
- 如果读和写分开，所有读操作都是串行的，看不出优势
- 只有读和写**同时进行**时，才能看到 RWMutex 允许多个读并发执行

---

### 4. 代码结构和逻辑（10/20）⚠️

**优点**：
- ✅ 代码结构清晰
- ✅ 有对比测试

**问题**：
- ❌ 缺少性能对比输出（没有计算加速比）
- ❌ 变量命名错误：`timeRear2` 应该是 `timeRead2`
- ❌ 测试场景不符合"多读少写同时进行"的要求

---

## 主要问题分析

### 问题1：测试方式不正确

**当前方式**（错误）：
```
1. 启动所有读 goroutine → 等待完成
2. 启动所有写 goroutine → 等待完成
3. 重复上述步骤测试 Mutex 版本
```

**正确方式**：
```
1. 同时启动读和写 goroutine（多读少写）
2. 等待所有完成
3. 对比 RWMutex 和 Mutex 的总耗时
```

### 问题2：缺少性能对比

当前代码只输出了各自的耗时，没有：
- 计算加速比
- 明确说明哪个更快
- 总结 RWMutex 的优势

---

## 改进建议

### 1. 修改测试逻辑（关键）

```go
func main() {
    const readerCount = 1000
    const writerCount = 10
    const operations = 1000

    // 测试1：RWMutex（读和写同时进行）
    fmt.Println("=== 测试1：RWMutex（读和写同时进行）===")
    cache1 := NewCache()
    cache1.Set("Key1", "Value1")
    cache1.Set("Key2", "Value2")

    var wg1 sync.WaitGroup
    start1 := time.Now()

    // 同时启动读和写的 goroutine
    // 读 goroutine
    for i := 0; i < readerCount; i++ {
        wg1.Add(1)
        go func() {
            defer wg1.Done()
            for j := 0; j < operations; j++ {
                cache1.Get("Key1")
                cache1.Get("Key2")
            }
        }()
    }

    // 写 goroutine（同时进行）
    for i := 0; i < writerCount; i++ {
        wg1.Add(1)
        go func(id int) {
            defer wg1.Done()
            for j := 0; j < operations/10; j++ {
                cache1.Set("Key1", fmt.Sprintf("Value1-%d", j))
                cache1.Set("Key2", fmt.Sprintf("Value2-%d", j))
            }
        }(i)
    }

    wg1.Wait()
    cost1 := time.Since(start1)
    fmt.Printf("RWMutex 总耗时: %v\n\n", cost1)

    // 测试2：Mutex（读和写同时进行）
    fmt.Println("=== 测试2：Mutex（读和写同时进行）===")
    cache2 := NewCacheWithMutex()
    cache2.Set("Key1", "Value1")
    cache2.Set("Key2", "Value2")

    var wg2 sync.WaitGroup
    start2 := time.Now()

    // 同时启动读和写的 goroutine
    for i := 0; i < readerCount; i++ {
        wg2.Add(1)
        go func() {
            defer wg2.Done()
            for j := 0; j < operations; j++ {
                cache2.Get("Key1")
                cache2.Get("Key2")
            }
        }()
    }

    for i := 0; i < writerCount; i++ {
        wg2.Add(1)
        go func(id int) {
            defer wg2.Done()
            for j := 0; j < operations/10; j++ {
                cache2.Set("Key1", fmt.Sprintf("Value1-%d", j))
                cache2.Set("Key2", fmt.Sprintf("Value2-%d", j))
            }
        }(i)
    }

    wg2.Wait()
    cost2 := time.Since(start2)
    fmt.Printf("Mutex 总耗时: %v\n\n", cost2)

    // 性能对比
    if cost1 < cost2 {
        speedup := float64(cost2) / float64(cost1)
        fmt.Printf("✅ RWMutex 比 Mutex 快 %.2fx（多读场景的优势）\n", speedup)
    } else {
        fmt.Println("⚠️ 在这个场景下，RWMutex 可能没有明显优势")
    }
}
```

### 2. 修复变量命名

```go
// 当前（错误）
timeRear2 := time.Now()

// 应该改为
timeRead2 := time.Now()
```

---

## 知识点总结

### ✅ 你掌握的内容

1. **RWMutex 基本用法**：
   - `RLock()` / `RUnlock()` 用于读操作
   - `Lock()` / `Unlock()` 用于写操作

2. **读写锁的概念**：
   - 读锁允许多个 goroutine 同时持有
   - 写锁是独占的

### ⚠️ 需要改进的地方

1. **测试场景理解**：
   - RWMutex 的优势在于"多读少写**同时进行**"
   - 需要同时启动读和写的 goroutine，而不是分开测试

2. **性能对比**：
   - 应该计算加速比，明确说明性能差异

---

## 总结

**总体评价**：⭐⭐⭐

**优点**：
- ✅ RWMutex 使用完全正确
- ✅ 代码结构清晰
- ✅ 有对比测试的意识

**需要改进**：
- ❌ 测试逻辑不符合"多读少写同时进行"的场景
- ❌ 缺少性能对比输出
- ⚠️ 变量命名有小错误

**核心问题**：测试方式需要修改，让读和写**同时进行**，这样才能真正体现出 RWMutex 的优势。

---

**建议**：修改测试逻辑后重新运行，应该能看到 RWMutex 在多读场景下的性能优势。

