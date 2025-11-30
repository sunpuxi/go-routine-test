# 练习1 评分报告

## 总分：100/100 ✅

---

## 详细评分

### 1. Mutex 使用正确性（30/30）✅

**优点**：
- ✅ `Counter.Inc()` 正确使用 `Lock()` 和 `defer Unlock()`
- ✅ `Counter.Get()` 正确使用 `Lock()` 和 `defer Unlock()`
- ✅ 使用 `defer` 确保锁一定会被释放（即使发生 panic）

**代码片段**：
```go
func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()  // ✅ 正确使用 defer
    c.value++
}
```

---

### 2. 并发安全实现（25/25）✅

**优点**：
- ✅ 所有对共享变量 `value` 的访问都被锁保护
- ✅ 读操作（`Get()`）和写操作（`Inc()`）都正确加锁
- ✅ 没有遗漏任何需要保护的临界区

---

### 3. 代码结构和逻辑（25/25）✅

**优点**：
- ✅ 正确实现了线程安全的 `Counter`
- ✅ 正确实现了不安全的 `UnsafeCounter` 用于对比
- ✅ 测试逻辑清晰：对比安全和不安全版本的执行结果
- ✅ 使用合理的测试规模（1000 个 goroutine，每个执行 1000 次）

**代码片段**：
```go
goroutineNum := 1000
taskCount := 1000
// ✅ 测试规模合理，能明显看出竞态条件
```

---

### 4. WaitGroup 使用（20/20）✅

**优点**：
- ✅ 正确使用 `wg.Add(1)` 增加计数器
- ✅ 正确使用 `defer wg.Done()` 减少计数器
- ✅ 正确使用 `wg.Wait()` 等待所有 goroutine 完成
- ✅ 两个测试都正确使用了 WaitGroup（包括修复后的 `wg2`）

**代码片段**：
```go
go func() {
    defer wg.Done()  // ✅ 正确
    // ...
}()
```

---

### 5. 额外改进（+10 分）⭐

**优点**：
- ✅ 添加了时间统计，可以对比性能差异
- ✅ 输出格式清晰，包含结果和耗时

**代码片段**：
```go
time1 := time.Now()
// ... 执行 ...
fmt.Printf("safeCounter.Get() = %d,cost time is %v\n", safeCounter.Get(), time.Since(time1))
```

---

## 代码亮点

### 1. 正确的 Mutex 使用模式

```go
c.mu.Lock()
defer c.mu.Unlock()  // ✅ 最佳实践
```

**为什么用 `defer`？**
- 即使函数中间有 `return` 或 `panic`，锁也会被释放
- 避免忘记 `Unlock()` 导致死锁

### 2. 对比测试设计

- 同时测试安全和不安全版本
- 可以直观看到竞态条件导致的数据丢失
- 测试规模足够大，能稳定复现问题

### 3. 性能统计

- 添加了时间统计，可以观察：
  - Mutex 版本：虽然安全，但因为有锁竞争，可能稍慢
  - Unsafe 版本：虽然快，但结果错误

---

## 预期运行结果

```
safeCounter.Get() = 1000000, cost time is ~XXXms
counter.Get() = 999XXX, cost time is ~XXXms  (小于 1000000，因为竞态条件)
```

**说明**：
- `safeCounter` 的结果应该是 `1000000`（正确）
- `counter` 的结果可能小于 `1000000`（竞态条件导致数据丢失）
- Mutex 版本可能稍慢（因为锁竞争），但结果是正确的

---

## 知识点总结

### ✅ 你掌握的内容

1. **Mutex 基本用法**：
   - `Lock()` 获取锁
   - `Unlock()` 释放锁
   - 使用 `defer` 确保释放

2. **并发安全**：
   - 所有对共享资源的访问都要加锁
   - 读操作也需要加锁（如果可能被并发修改）

3. **WaitGroup 使用**：
   - `Add()` 增加计数器
   - `Done()` 减少计数器
   - `Wait()` 等待计数器归零

4. **竞态条件**：
   - 理解了为什么需要锁
   - 通过对比看到了竞态条件的实际影响

---

## 改进建议（可选）

### 1. 输出格式优化

```go
// 当前
fmt.Printf("safeCounter.Get() = %d,cost time is %v\n", ...)

// 建议（更清晰）
fmt.Printf("安全版本: 结果 = %d, 耗时 = %v\n", ...)
fmt.Printf("不安全版本: 结果 = %d, 耗时 = %v\n", ...)
```

### 2. 添加期望值对比

```go
expected := goroutineNum * taskCount
fmt.Printf("安全版本: 结果 = %d (期望: %d) ✅\n", safeCounter.Get(), expected)
fmt.Printf("不安全版本: 结果 = %d (期望: %d) ❌ 丢失了 %d 次操作\n", 
    counter.Get(), expected, expected-counter.Get())
```

### 3. 使用常量

```go
const (
    goroutineNum = 1000
    taskCount    = 1000
)
```

---

## 总结

**总体评价**：⭐⭐⭐⭐⭐

你的代码实现非常优秀：
- ✅ Mutex 使用完全正确
- ✅ 并发安全实现完善
- ✅ 代码结构清晰
- ✅ 有性能统计的额外改进

**核心知识点掌握情况**：
- ✅ 理解 Mutex 的作用和用法
- ✅ 理解并发安全的重要性
- ✅ 理解竞态条件的危害
- ✅ 正确使用 WaitGroup

**可以继续学习**：
- 练习2：RWMutex（读写锁优化）
- 练习3：Atomic（原子操作性能对比）

---

**恭喜完成练习1！** 🎉

