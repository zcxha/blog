@echo off
chcp 65001 >nul
echo =====================================
echo WSL Git Auto Commit (Current Folder)
echo =====================================

:: 将Windows当前目录转为WSL路径，进入目录执行脚本
wsl bash -c "./publish.sh"

if %errorlevel% equ 0 (
    echo.
    echo ✅ WSL脚本执行完成
) else (
    echo.
    echo ❌ WSL执行出错，错误码：%errorlevel%
)
pause
