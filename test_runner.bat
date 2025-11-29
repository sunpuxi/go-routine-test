@echo off
echo 分布式锁测试工具
echo ==================

echo.
echo 1. 并发测试 - 启动多个进程同时获取锁
echo 2. 安全性测试 - 测试锁的安全机制
echo 3. 超时测试 - 测试锁的自动过期
echo 4. 退出
echo.

set /p choice=请选择测试类型 (1-4): 

if "%choice%"=="1" goto concurrent_test
if "%choice%"=="2" goto safety_test
if "%choice%"=="3" goto timeout_test
if "%choice%"=="4" goto end
goto invalid

:concurrent_test
echo.
echo 启动并发测试...
echo 将启动3个进程同时尝试获取锁
echo.
start "进程1" cmd /k "go run test_lock.go process1 concurrent"
timeout /t 2 /nobreak >nul
start "进程2" cmd /k "go run test_lock.go process2 concurrent"
timeout /t 2 /nobreak >nul
start "进程3" cmd /k "go run test_lock.go process3 concurrent"
goto end

:safety_test
echo.
echo 启动安全性测试...
echo 将启动2个进程测试锁的安全性
echo.
start "安全测试1" cmd /k "go run test_lock.go safety1 safety"
timeout /t 2 /nobreak >nul
start "安全测试2" cmd /k "go run test_lock.go safety2 safety"
goto end

:timeout_test
echo.
echo 启动超时测试...
echo 将启动1个进程测试锁的自动过期
echo.
start "超时测试" cmd /k "go run test_lock.go timeout1 timeout"
goto end

:invalid
echo 无效选择，请重新运行脚本
goto end

:end
echo.
echo 测试已启动，请观察各个窗口的输出结果
pause
