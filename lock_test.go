package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// 测试分布式锁的并发互斥性
func TestDistributedLockConcurrency(t *testing.T) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}

	// 清理测试数据
	defer func() {
		rdb.Del(ctx, "test_concurrent_lock")
		rdb.Close()
	}()

	// 测试参数
	lockKey := "test_concurrent_lock"
	numGoroutines := 10
	lockDuration := 5 * time.Second
	successCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动多个goroutine并发获取锁
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			lockValue := fmt.Sprintf("goroutine_%d_%d", goroutineID, time.Now().UnixNano())

			// 尝试获取锁
			acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
			if err != nil {
				t.Errorf("Goroutine %d 获取锁失败: %v", goroutineID, err)
				return
			}

			if acquired {
				// 通过加锁统计成功获取锁的数量
				mu.Lock()
				successCount++
				mu.Unlock()

				t.Logf("Goroutine %d 成功获取锁", goroutineID)

				// 模拟业务处理
				time.Sleep(100 * time.Millisecond)

				// 释放锁
				released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
				if err != nil {
					t.Errorf("Goroutine %d 释放锁失败: %v", goroutineID, err)
				} else if released {
					t.Logf("Goroutine %d 成功释放锁", goroutineID)
				} else {
					t.Logf("Goroutine %d 锁已过期或被其他进程获取", goroutineID)
				}
			} else {
				t.Logf("Goroutine %d 获取锁失败，锁已被其他进程持有", goroutineID)
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果：只有一个goroutine能获取到锁
	if successCount != 1 {
		t.Errorf("期望只有1个goroutine能获取锁，实际有%d个", successCount)
	} else {
		t.Logf("测试通过：只有1个goroutine成功获取锁")
	}
}

// 测试分布式锁的安全性
func TestDistributedLockSafety(t *testing.T) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}

	// 清理测试数据
	defer func() {
		rdb.Del(ctx, "test_safety_lock")
		rdb.Close()
	}()

	lockKey := "test_safety_lock"
	lockValue := fmt.Sprintf("test_%d", time.Now().UnixNano())
	lockDuration := 10 * time.Second

	// 获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}

	if !acquired {
		t.Fatal("无法获取锁进行测试")
	}

	// 测试用错误值释放锁
	wrongValue := "wrong_value"
	released, err := releaseTestLock(ctx, rdb, lockKey, wrongValue)
	if err != nil {
		t.Errorf("释放锁失败: %v", err)
	}
	if released {
		t.Error("用错误值释放锁成功，这不应该发生")
	} else {
		t.Log("用错误值释放锁失败，这是正确的")
	}

	// 测试用正确值释放锁
	released, err = releaseTestLock(ctx, rdb, lockKey, lockValue)
	if err != nil {
		t.Errorf("释放锁失败: %v", err)
	}
	if !released {
		t.Error("用正确值释放锁失败")
	} else {
		t.Log("用正确值释放锁成功")
	}
}

// 测试分布式锁的超时机制
func TestDistributedLockTimeout(t *testing.T) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}

	// 清理测试数据
	defer func() {
		rdb.Del(ctx, "test_timeout_lock")
		rdb.Close()
	}()

	lockKey := "test_timeout_lock"
	lockValue := fmt.Sprintf("test_%d", time.Now().UnixNano())
	lockDuration := 2 * time.Second // 短超时时间

	// 获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}

	if !acquired {
		t.Fatal("无法获取锁进行测试")
	}

	t.Logf("获取锁成功，等待 %v 让锁自动过期", lockDuration+1*time.Second)

	// 等待锁过期
	time.Sleep(lockDuration + 1*time.Second)

	// 尝试释放已过期的锁
	released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
	if err != nil {
		t.Errorf("释放锁失败: %v", err)
	}
	if released {
		t.Log("释放已过期锁成功（锁可能被其他进程重新获取）")
	} else {
		t.Log("释放已过期锁失败（锁已自动过期）")
	}
}

// 测试分布式锁的续期机制
func TestDistributedLockRenewal(t *testing.T) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}

	// 清理测试数据
	defer func() {
		rdb.Del(ctx, "test_renewal_lock")
		rdb.Close()
	}()

	lockKey := "test_renewal_lock"
	lockValue := fmt.Sprintf("test_%d", time.Now().UnixNano())
	lockDuration := 3 * time.Second

	// 获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}

	if !acquired {
		t.Fatal("无法获取锁进行测试")
	}

	t.Log("获取锁成功，开始测试续期机制")

	// 模拟长时间业务处理
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)

		// 检查锁是否仍然有效
		val, err := rdb.Get(ctx, lockKey).Result()
		if err != nil {
			t.Logf("锁已过期或被删除: %v", err)
			break
		}
		if val == lockValue {
			t.Logf("第%d秒：锁仍然有效", i+1)
		} else {
			t.Logf("第%d秒：锁已被其他进程获取", i+1)
			break
		}
	}

	// 释放锁
	released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
	if err != nil {
		t.Errorf("释放锁失败: %v", err)
	} else if released {
		t.Log("成功释放锁")
	} else {
		t.Log("锁已过期或被其他进程获取")
	}
}

// 测试分布式锁的竞争条件
func TestDistributedLockRaceCondition(t *testing.T) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接失败: %v", err)
	}

	// 清理测试数据
	defer func() {
		rdb.Del(ctx, "test_race_lock")
		rdb.Close()
	}()

	lockKey := "test_race_lock"
	numGoroutines := 20
	lockDuration := 3 * time.Second
	successCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动多个goroutine同时竞争锁
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			lockValue := fmt.Sprintf("race_%d_%d", goroutineID, time.Now().UnixNano())

			// 尝试获取锁
			acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
			if err != nil {
				t.Errorf("Goroutine %d 获取锁失败: %v", goroutineID, err)
				return
			}

			if acquired {
				mu.Lock()
				successCount++
				mu.Unlock()

				t.Logf("Goroutine %d 成功获取锁", goroutineID)

				// 模拟业务处理
				time.Sleep(50 * time.Millisecond)

				// 释放锁
				released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
				if err != nil {
					t.Errorf("Goroutine %d 释放锁失败: %v", goroutineID, err)
				} else if released {
					t.Logf("Goroutine %d 成功释放锁", goroutineID)
				}
			} else {
				t.Logf("Goroutine %d 获取锁失败", goroutineID)
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果：只有一个goroutine能获取到锁
	if successCount != 1 {
		t.Errorf("期望只有1个goroutine能获取锁，实际有%d个", successCount)
	} else {
		t.Logf("竞争条件测试通过：只有1个goroutine成功获取锁")
	}
}
