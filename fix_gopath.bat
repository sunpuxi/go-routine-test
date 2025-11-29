@echo off
echo 修复GOPATH警告
echo ================

echo 设置环境变量...
set GO111MODULE=on
set GOPATH=

echo 运行测试...
go test -v

echo.
echo 测试完成！
pause
