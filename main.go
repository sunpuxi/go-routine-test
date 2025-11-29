package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis地址
		Password: "",               // 密码
		DB:       0,                // 数据库
	})

	ctx := context.Background()

	// 测试连接，带重试机制
	var pong string
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		pong, err = rdb.Ping(ctx).Result()
		if err == nil {
			break
		}
		fmt.Printf("Redis连接失败，重试 %d/%d: %v\n", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Fatal("Redis连接失败，请确保Redis服务在6479端口运行:", err)
	}
	fmt.Println("Redis连接成功:", pong)

	// 基本操作示例
	basicOperations(ctx, rdb)

	// 分布式锁示例
	distributedLockExample(ctx, rdb)

	// 关闭连接
	rdb.Close()
}

// 基本Redis操作
func basicOperations(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== 基本Redis操作 ===")

	// SET操作
	err := rdb.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		log.Printf("SET操作失败: %v", err)
		return
	}
	fmt.Println("SET key1 value1 成功")

	// GET操作
	val, err := rdb.Get(ctx, "key1").Result()
	if err != nil {
		log.Printf("GET操作失败: %v", err)
		return
	}
	fmt.Printf("GET key1: %s\n", val)

	// 设置过期时间
	err = rdb.Set(ctx, "key2", "value2", 10*time.Second).Err()
	if err != nil {
		log.Printf("SET with TTL失败: %v", err)
		return
	}
	fmt.Println("SET key2 value2 (10秒过期) 成功")

	// 检查键是否存在
	exists, err := rdb.Exists(ctx, "key1").Result()
	if err != nil {
		log.Printf("EXISTS操作失败: %v", err)
		return
	}
	fmt.Printf("key1存在: %d\n", exists)

	// 删除键
	err = rdb.Del(ctx, "key1").Err()
	if err != nil {
		log.Printf("DEL操作失败: %v", err)
		return
	}
	fmt.Println("DEL key1 成功")
}

// 分布式锁示例
func distributedLockExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== 分布式锁示例 ===")

	lockKey := "mylock"
	lockValue := "lock_value_123"
	lockDuration := 30 * time.Second

	// 获取锁
	acquired, err := acquireLock(ctx, rdb, lockKey, lockValue, lockDuration)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	if acquired {
		fmt.Println("成功获取分布式锁")

		// 模拟业务处理
		time.Sleep(2 * time.Second)

		// 释放锁
		released, err := releaseLock(ctx, rdb, lockKey, lockValue)
		if err != nil {
			log.Printf("释放锁失败: %v", err)
		} else if released {
			fmt.Println("成功释放分布式锁")
		} else {
			fmt.Println("锁已过期或被其他进程获取")
		}
	} else {
		fmt.Println("获取锁失败，锁已被其他进程持有")
	}
}

// 获取分布式锁
func acquireLock(ctx context.Context, rdb *redis.Client, key, value string, duration time.Duration) (bool, error) {
	result, err := rdb.SetNX(ctx, key, value, duration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// 释放分布式锁
func releaseLock(ctx context.Context, rdb *redis.Client, key, value string) (bool, error) {
	// 使用Lua脚本确保原子性：只有锁的持有者才能释放锁
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
