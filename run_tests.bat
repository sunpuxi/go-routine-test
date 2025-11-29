@echo off
echo 分布式锁单元测试
echo ==================

echo.
echo 1. 运行所有测试
echo 2. 运行并发测试
echo 3. 运行安全性测试
echo 4. 运行超时测试
echo 5. 运行续期测试
echo 6. 运行竞争条件测试
echo 7. 退出
echo.

set /p choice=请选择测试类型 (1-7): 

if "%choice%"=="1" goto all_tests
if "%choice%"=="2" goto concurrent_test
if "%choice%"=="3" goto safety_test
if "%choice%"=="4" goto timeout_test
if "%choice%"=="5" goto renewal_test
if "%choice%"=="6" goto race_test
if "%choice%"=="7" goto end
goto invalid

:all_tests
echo.
echo 运行所有测试...
go test -v -run TestDistributedLock
goto end

:concurrent_test
echo.
echo 运行并发测试...
go test -v -run TestDistributedLockConcurrency
goto end

:safety_test
echo.
echo 运行安全性测试...
go test -v -run TestDistributedLockSafety
goto end

:timeout_test
echo.
echo 运行超时测试...
go test -v -run TestDistributedLockTimeout
goto end

:renewal_test
echo.
echo 运行续期测试...
go test -v -run TestDistributedLockRenewal
goto end

:race_test
echo.
echo 运行竞争条件测试...
go test -v -run TestDistributedLockRaceCondition
goto end

:invalid
echo 无效选择，请重新运行脚本
goto end

:end
echo.
echo 测试完成
pause
