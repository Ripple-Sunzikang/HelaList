@echo off
setlocal

:: --- 配置 ---
set "USERNAME=suzuki"
set "PASSWORD=suzuki"
set "LOGIN_URL=http://localhost:8080/api/user/login"
set "LIST_URL=http://localhost:8080/api/fs/list/localwebdav?refresh=true"
set "TEMP_FILE=login_response.json"

:: 1. 登录并获取token，将响应保存到临时文件
echo 正在登录到 %LOGIN_URL% ...
curl -s -X POST %LOGIN_URL% ^
    -H "Content-Type: application/json" ^
    -d "{\"username\":\"%USERNAME%\",\"password\":\"%PASSWORD%\"}" > %TEMP_FILE%

:: 2. 使用PowerShell从临时文件中解析出登录状态码
:: 注意: 此脚本需要PowerShell (Windows 7及以上版本均自带)
for /f "usebackq tokens=*" %%i in (`powershell -Command "(Get-Content -Raw -Path %TEMP_FILE% | ConvertFrom-Json).code"`) do (
    set "LOGIN_CODE=%%i"
)

:: 3. 判断登录是否成功
if "%LOGIN_CODE%"=="200" (
    echo 登录成功，正在提取token...
    
    :: 登录成功，同样用PowerShell提取token
    for /f "usebackq tokens=*" %%j in (`powershell -Command "(Get-Content -Raw -Path %TEMP_FILE% | ConvertFrom-Json).data.token"`) do (
        set "TOKEN=%%j"
    )
    echo 获取到token: %TOKEN%
    echo.
    echo --- 文件列表 ---
    
    :: 4. 使用获取到的token查询文件列表
    curl -s -X GET %LIST_URL% ^
        -H "Content-Type: application/json" ^
        -H "Authorization: Bearer %TOKEN%"
    echo.

) else (
    echo.
    echo [!] 登录失败，服务器响应:
    type %TEMP_FILE%
    echo.
)

:: 5. 清理临时文件
if exist %TEMP_FILE% (
    del %TEMP_FILE%
)

endlocal
pause