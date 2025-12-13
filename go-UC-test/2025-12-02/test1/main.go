package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// 练习1（简单）：sync.Once 实现单例模式和延迟初始化
// 场景：某些资源只需要初始化一次，即使多个 goroutine 并发调用也应该只执行一次
// 要求：
// 1）使用 sync.Once 实现一个线程安全的单例模式
// 2）对比不使用 Once 的版本，观察多次初始化的问题
// 3）理解 Once.Do() 的语义：只执行一次，即使多个 goroutine 同时调用
// 4）实现一个延迟初始化的配置管理器，多个 goroutine 可以安全地获取配置
// ============================================

// ============ 示例1：不使用 Once 的单例（有问题） ============

// SingletonBad 非线程安全的单例
type SingletonBad struct {
	value string
}

var instanceBad *SingletonBad
var initCountBad int // 用于统计初始化次数

// GetInstanceBad 非线程安全的单例获取方法
func GetInstanceBad() *SingletonBad {
	if instanceBad == nil {
		// 模拟初始化耗时
		time.Sleep(10 * time.Millisecond)
		instanceBad = &SingletonBad{value: "单例实例"}
		initCountBad++
	}
	return instanceBad
}

// ============ 示例2：使用 Once 的单例（正确） ============

// SingletonGood 线程安全的单例
type SingletonGood struct {
	value string
}

var (
	instanceGood *SingletonGood
	once         sync.Once
	initCountGood int
)

// GetInstanceGood 线程安全的单例获取方法
func GetInstanceGood() *SingletonGood {
	once.Do(func() {
		// 模拟初始化耗时
		time.Sleep(10 * time.Millisecond)
		instanceGood = &SingletonGood{value: "单例实例"}
		initCountGood++
	})
	return instanceGood
}

// ============ 示例3：延迟初始化的配置管理器 ============

// Config 配置结构
type Config struct {
	DatabaseURL string
	APIKey      string
	Port        int
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config *Config
	once   sync.Once
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// GetConfig 获取配置（延迟初始化）
func (cm *ConfigManager) GetConfig() *Config {
	cm.once.Do(func() {
		// 模拟从文件或数据库加载配置的耗时操作
		fmt.Println("正在初始化配置...")
		time.Sleep(50 * time.Millisecond)
		cm.config = &Config{
			DatabaseURL: "localhost:5432",
			APIKey:      "secret-key-123",
			Port:        8080,
		}
		fmt.Println("配置初始化完成")
	})
	return cm.config
}

// ============ 测试函数 ============

func testBadSingleton() {
	fmt.Println("\n=== 测试：不使用 Once 的单例（有问题） ===")
	instanceBad = nil
	initCountBad = 0
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// 启动多个 goroutine 并发获取单例
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			instance := GetInstanceBad()
			_ = instance // 使用实例
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("初始化次数: %d (期望为1，实际可能有多次)\n", initCountBad)
}

func testGoodSingleton() {
	fmt.Println("\n=== 测试：使用 Once 的单例（正确） ===")
	instanceGood = nil
	initCountGood = 0
	once = sync.Once{} // 重置 Once
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// 启动多个 goroutine 并发获取单例
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			instance := GetInstanceGood()
			_ = instance // 使用实例
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("初始化次数: %d (期望为1)\n", initCountGood)
}

func testConfigManager() {
	fmt.Println("\n=== 测试：延迟初始化的配置管理器 ===")
	manager := NewConfigManager()
	
	var wg sync.WaitGroup
	numGoroutines := 50
	
	// 多个 goroutine 同时获取配置
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			config := manager.GetConfig()
			fmt.Printf("Goroutine %d 获取配置: Port=%d\n", id, config.Port)
		}(i)
	}
	
	wg.Wait()
	fmt.Println("所有 goroutine 都获取到了配置")
}

func main() {
	fmt.Println("【专题5 - 练习1：sync.Once 实现单例模式和延迟初始化】")
	
	// 测试1：不使用 Once 的单例（会多次初始化）
	testBadSingleton()
	
	// 测试2：使用 Once 的单例（只初始化一次）
	testGoodSingleton()
	
	// 测试3：延迟初始化的配置管理器
	testConfigManager()
	
	fmt.Println("\n=== 练习提示 ===")
	fmt.Println("1. sync.Once 确保某个函数只执行一次，即使多个 goroutine 同时调用")
	fmt.Println("2. Once.Do() 是阻塞的，会等待初始化完成")
	fmt.Println("3. Once 是不可复用的，如果需要重新初始化，需要创建新的 Once 实例")
	fmt.Println("4. 适用于：单例模式、延迟初始化、全局资源初始化等场景")
}

