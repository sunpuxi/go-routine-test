# 练习1：基础 Worker Pool（中等难度）

## 题目描述

实现一个基础的 Worker Pool，用于并发处理任务。

## 需求

1. **创建固定数量的 Worker**：
   - Worker 数量可配置（例如 3 个）
   - 每个 Worker 从任务 channel 中取任务并处理

2. **任务处理**：
   - 任务是一个整数（代表任务ID）
   - 处理任务：计算 `taskID * taskID`，并打印结果
   - 模拟耗时：每个任务处理需要 100ms

3. **任务分发**：
   - 主 goroutine 负责发送 10 个任务（ID: 0-9）到任务 channel
   - 发送完所有任务后，关闭任务 channel

4. **结果收集**：
   - 每个 Worker 处理完任务后，将结果发送到结果 channel
   - 主 goroutine 收集所有结果并打印

5. **正确关闭**：
   - 所有 Worker 处理完任务后正常退出
   - 程序正常结束，无死锁

## 输入输出示例

```
Worker 0: 开始处理任务 0
Worker 1: 开始处理任务 1
Worker 2: 开始处理任务 2
Worker 0: 任务 0 完成，结果 = 0
Worker 1: 任务 1 完成，结果 = 1
Worker 2: 任务 2 完成，结果 = 1
Worker 0: 开始处理任务 3
Worker 1: 开始处理任务 4
...
主 goroutine: 收到结果 0
主 goroutine: 收到结果 1
主 goroutine: 收到结果 1
...
主 goroutine: 所有任务完成
```

## 评分标准

- ✅ 正确实现 Worker Pool 模式（20分）
- ✅ 正确使用 channel 和 goroutine（20分）
- ✅ 正确关闭 channel，无死锁（20分）
- ✅ 正确使用 WaitGroup 等待所有 Worker 完成（20分）
- ✅ 代码结构清晰，注释合理（20分）

## 提示

- 使用 `sync.WaitGroup` 等待所有 Worker 完成
- 记得关闭任务 channel 和结果 channel
- Worker 使用 `for range` 从任务 channel 取任务

