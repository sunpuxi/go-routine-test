package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func testLockMain() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_lock.go <进程ID> [测试类型]")
		fmt.Println("测试类型: concurrent, safety, timeout")
		return
	}

	processID := os.Args[1]
	testType := "concurrent"
	if len(os.Args) > 2 {
		testType = os.Args[2]
	}

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
		log.Fatal("Redis连接失败:", err)
	}

	fmt.Printf("进程 %s 开始测试分布式锁 (测试类型: %s)\n", processID, testType)

	switch testType {
	case "concurrent":
		testConcurrentLock(ctx, rdb, processID)
	case "safety":
		testLockSafety(ctx, rdb, processID)
	case "timeout":
		testLockTimeout(ctx, rdb, processID)
	default:
		fmt.Println("未知的测试类型:", testType)
	}

	rdb.Close()
}

// 测试并发获取锁
func testConcurrentLock(ctx context.Context, rdb *redis.Client, processID string) {
	lockKey := "test_concurrent_lock"
	lockValue := fmt.Sprintf("process_%s_%d", processID, time.Now().UnixNano())
	lockDuration := 10 * time.Second

	fmt.Printf("进程 %s 尝试获取锁...\n", processID)

	// 尝试获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	if acquired {
		fmt.Printf("进程 %s 成功获取锁，开始执行业务逻辑...\n", processID)

		// 模拟业务处理
		for i := 1; i <= 5; i++ {
			fmt.Printf("进程 %s 正在处理业务 %d/5\n", processID, i)
			time.Sleep(1 * time.Second)
		}

		// 释放锁
		released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
		if err != nil {
			log.Printf("释放锁失败: %v", err)
		} else if released {
			fmt.Printf("进程 %s 成功释放锁\n", processID)
		} else {
			fmt.Printf("进程 %s 锁已过期或被其他进程获取\n", processID)
		}
	} else {
		fmt.Printf("进程 %s 获取锁失败，锁已被其他进程持有\n", processID)
	}
}

// 测试锁的安全性
func testLockSafety(ctx context.Context, rdb *redis.Client, processID string) {
	lockKey := "test_safety_lock"
	lockValue := fmt.Sprintf("process_%s_%d", processID, time.Now().UnixNano())
	lockDuration := 15 * time.Second

	fmt.Printf("进程 %s 开始安全性测试...\n", processID)

	// 获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	if acquired {
		fmt.Printf("进程 %s 获取锁成功\n", processID)

		// 尝试用错误的值释放锁
		wrongValue := "wrong_value"
		fmt.Printf("进程 %s 尝试用错误值释放锁...\n", processID)
		released, err := releaseTestLock(ctx, rdb, lockKey, wrongValue)
		if err != nil {
			log.Printf("释放锁失败: %v", err)
		} else if released {
			fmt.Printf("进程 %s 用错误值释放锁成功 (这不应该发生!)\n", processID)
		} else {
			fmt.Printf("进程 %s 用错误值释放锁失败 (这是正确的)\n", processID)
		}

		// 用正确的值释放锁
		fmt.Printf("进程 %s 尝试用正确值释放锁...\n", processID)
		released, err = releaseTestLock(ctx, rdb, lockKey, lockValue)
		if err != nil {
			log.Printf("释放锁失败: %v", err)
		} else if released {
			fmt.Printf("进程 %s 用正确值释放锁成功\n", processID)
		} else {
			fmt.Printf("进程 %s 用正确值释放锁失败\n", processID)
		}
	} else {
		fmt.Printf("进程 %s 获取锁失败\n", processID)
	}
}

// 测试锁超时机制
func testLockTimeout(ctx context.Context, rdb *redis.Client, processID string) {
	lockKey := "test_timeout_lock"
	lockValue := fmt.Sprintf("process_%s_%d", processID, time.Now().UnixNano())
	lockDuration := 5 * time.Second // 短超时时间

	fmt.Printf("进程 %s 开始超时测试 (锁超时时间: %v)...\n", processID, lockDuration)

	// 获取锁
	acquired, err := acquireTestLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	if acquired {
		fmt.Printf("进程 %s 获取锁成功，将等待 %v 让锁自动过期...\n", processID, lockDuration+2*time.Second)

		// 等待锁过期
		time.Sleep(lockDuration + 2*time.Second)

		// 尝试释放已过期的锁
		fmt.Printf("进程 %s 尝试释放已过期的锁...\n", processID)
		released, err := releaseTestLock(ctx, rdb, lockKey, lockValue)
		if err != nil {
			log.Printf("释放锁失败: %v", err)
		} else if released {
			fmt.Printf("进程 %s 释放已过期锁成功 (锁可能被其他进程重新获取)\n", processID)
		} else {
			fmt.Printf("进程 %s 释放已过期锁失败 (锁已自动过期)\n", processID)
		}
	} else {
		fmt.Printf("进程 %s 获取锁失败\n", processID)
	}
}

// 获取分布式锁
func acquireTestLock(ctx context.Context, rdb *redis.Client, key, value string, duration time.Duration) (bool, error) {
	result, err := rdb.SetNX(ctx, key, value, duration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// 释放分布式锁
func releaseTestLock(ctx context.Context, rdb *redis.Client, key, value string) (bool, error) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := rdb.Eval(ctx, script, []string{key}, value).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
