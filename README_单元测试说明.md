# 分布式锁单元测试说明

## 测试概述

本测试套件专门验证Redis分布式锁在并发场景下的正确性和可靠性，确保只有一个线程能够获取到分布式锁。

## 测试用例

### 1. TestDistributedLockConcurrency - 并发互斥性测试
**目的**：验证多个goroutine同时竞争锁时，只有一个能成功获取

**测试场景**：
- 启动10个goroutine同时尝试获取同一个锁
- 验证只有一个goroutine能成功获取锁
- 验证其他goroutine获取锁失败

**预期结果**：
- 只有1个goroutine成功获取锁
- 其他9个goroutine获取锁失败
- 成功获取锁的goroutine能正常释放锁

### 2. TestDistributedLockSafety - 安全性测试
**目的**：验证锁的安全机制，防止误删

**测试场景**：
- 获取锁后，用错误的值尝试释放锁
- 用正确的值释放锁

**预期结果**：
- 用错误值释放锁失败
- 用正确值释放锁成功

### 3. TestDistributedLockTimeout - 超时机制测试
**目的**：验证锁的自动过期机制

**测试场景**：
- 获取锁后等待锁自动过期
- 尝试释放已过期的锁

**预期结果**：
- 锁在指定时间后自动过期
- 释放已过期的锁失败

### 4. TestDistributedLockRenewal - 续期机制测试
**目的**：验证锁在业务处理期间的有效性

**测试场景**：
- 获取锁后模拟长时间业务处理
- 定期检查锁是否仍然有效

**预期结果**：
- 锁在业务处理期间保持有效
- 锁在超时后自动过期

### 5. TestDistributedLockRaceCondition - 竞争条件测试
**目的**：验证高并发场景下的锁竞争

**测试场景**：
- 启动20个goroutine同时竞争锁
- 验证在高并发下的互斥性

**预期结果**：
- 只有1个goroutine成功获取锁
- 其他19个goroutine获取锁失败

## 运行测试

### 方法1：使用测试脚本
```bash
# Windows
run_tests.bat

# 选择测试类型：
# 1. 运行所有测试
# 2. 运行并发测试
# 3. 运行安全性测试
# 4. 运行超时测试
# 5. 运行续期测试
# 6. 运行竞争条件测试
```

### 方法2：命令行运行
```bash
# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestDistributedLockConcurrency
go test -v -run TestDistributedLockSafety
go test -v -run TestDistributedLockTimeout
go test -v -run TestDistributedLockRenewal
go test -v -run TestDistributedLockRaceCondition
```

## 测试结果解读

### 成功指标
- **并发测试**：只有1个goroutine成功获取锁
- **安全性测试**：错误值无法释放锁，正确值能释放锁
- **超时测试**：锁在指定时间后自动过期
- **续期测试**：锁在业务处理期间保持有效
- **竞争测试**：高并发下仍只有1个goroutine成功

### 失败情况
- 多个goroutine同时获取锁成功
- 错误值能释放锁
- 锁超时机制失效
- 高并发下锁竞争失败

## 测试环境要求

1. **Redis服务**：确保Redis在localhost:6379运行
2. **Go环境**：Go 1.16+
3. **依赖库**：github.com/go-redis/redis/v8

## 测试数据清理

每个测试都会自动清理测试数据，确保测试之间不会相互影响。

## 性能指标

- **并发测试**：10个goroutine，100ms业务处理时间
- **竞争测试**：20个goroutine，50ms业务处理时间
- **超时测试**：2秒锁超时时间
- **续期测试**：3秒锁超时时间，5秒业务处理

## 常见问题

### Q: 测试失败，显示多个goroutine获取锁成功？
A: 检查Redis连接和锁的实现逻辑，确保SetNX操作正确

### Q: 安全性测试失败？
A: 检查Lua脚本是否正确实现，确保只有锁持有者能释放锁

### Q: 超时测试失败？
A: 检查Redis的TTL设置是否正确

### Q: 测试运行缓慢？
A: 这是正常的，超时测试需要等待锁过期，续期测试需要长时间运行

## 测试覆盖范围

- ✅ 并发互斥性
- ✅ 锁安全性
- ✅ 锁超时机制
- ✅ 锁续期机制
- ✅ 高并发竞争
- ✅ 错误处理
- ✅ 资源清理

这些测试确保了分布式锁在各种场景下的正确性和可靠性。
