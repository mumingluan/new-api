# Xuancat 原版风格主页部署

本文说明如何在原版主页中启用 Xuancat 密钥工具，构建 Windows/Linux amd64 NewAPI 二进制，并部署到当前生产拓扑。关于页始终使用原版实现。

## 部署目标

- Windows 主节点：`C:\new-api\new-api.exe`，由 `MumlNewApi` 服务启动。
- NekoMetal 备用节点：`/opt/new-api/new-api`，由 `new-api.service` 启动。

部署只替换 NewAPI 可执行文件，不修改 `.env`、日志或服务配置，也不创建备份。

Xuancat 页面通过当前 NewAPI 实例访问：

- `/api/activation/*`
- `/api/activation-code/*`
- `/v1/dashboard/billing/*`
- `/api/log/token`

## 启用 Xuancat 模式

构建前设置：

```powershell
$env:VITE_LANDING_PAGE_VARIANT = 'xuancat'
```

`xuancat` 保留原版主页的 Hero、API 调用示例、统计、功能、使用流程和 CTA，仅在 Hero 下方增加密钥开通、密钥查询和 API 基础地址工具；`default` 或未设置该变量时使用完全原版主页。关于页不受此开关影响。

实现位置：

- 构建参数读取：`web/rsbuild.config.ts`
- 前端功能判断：`web/src/lib/landing-page-variant.ts`
- 原版 Hero 与工具入口：`web/src/features/home/components/sections/hero.tsx`
- Xuancat 密钥工具：`web/src/features/xuancat-pages/components/home-access-panel.tsx`
- 激活码管理：`web/src/features/activation-codes/`
- Docker 构建参数：根目录 `Dockerfile`

该开关是编译时开关。NewAPI 使用 `go:embed web/dist`，因此必须先构建前端，再编译 Go 二进制。

## 编译前端和 NewAPI

```powershell
Set-Location D:\qiqi-api\web
$env:VITE_LANDING_PAGE_VARIANT = 'xuancat'
bun install --frozen-lockfile
bun run typecheck
bun test
bun run build

Set-Location D:\qiqi-api
$version = (Get-Content VERSION -Raw).Trim()
if (-not $version) {
  $version = git describe --tags --always
}

$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -trimpath -ldflags "-s -w -X github.com/QuantumNous/new-api/common.Version=$version" -o bin/new-api-xuancat-windows-amd64.exe .

$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
go build -trimpath -ldflags "-s -w -X github.com/QuantumNous/new-api/common.Version=$version" -o bin/new-api-xuancat-linux-amd64 .
```

## 部署 Windows NewAPI

以下命令直接覆盖现有文件，不创建备份：

```powershell
Stop-Service MumlNewApi -Force
$stopDeadline = (Get-Date).AddSeconds(30)
do {
  Start-Sleep -Milliseconds 500
  $service = Get-CimInstance Win32_Service -Filter "Name='MumlNewApi'"
} while ($service.State -ne 'Stopped' -and (Get-Date) -lt $stopDeadline)

Get-CimInstance Win32_Process |
  Where-Object { $_.ExecutablePath -eq 'C:\new-api\new-api.exe' } |
  ForEach-Object { Stop-Process -Id $_.ProcessId -Force }

Copy-Item -LiteralPath D:\qiqi-api\bin\new-api-xuancat-windows-amd64.exe `
  -Destination C:\new-api\new-api.exe -Force
Start-Service MumlNewApi

Invoke-WebRequest -UseBasicParsing http://127.0.0.1:65477/api/status
```

## 部署 NekoMetal NewAPI

```powershell
scp D:\qiqi-api\bin\new-api-xuancat-linux-amd64 NekoMetal:/opt/new-api/new-api.new
ssh NekoMetal 'set -e
  chmod 0755 /opt/new-api/new-api.new
  systemctl stop new-api
  mv -f /opt/new-api/new-api.new /opt/new-api/new-api
  systemctl start new-api
  systemctl is-active new-api
  curl -fsS http://127.0.0.1:65477/api/status >/dev/null'
```

部署完成后比较本地构建产物与目标文件的 SHA-256，并删除本地 `bin/new-api-xuancat-*` 产物；目标目录中不保留 `.new` 或备份文件。

## 发布后检查

1. Windows `MumlNewApi` 状态为 `Running`，本机 `/api/status` 返回 HTTP 200。
2. NekoMetal `new-api.service` 状态为 `active`，远端 `/api/status` 返回 HTTP 200。
3. 两个目标的二进制哈希分别与对应构建产物一致。
4. `/457` 返回 `{"457":true}`。
5. Xuancat 主页保留原版 Hero/API 示例，密钥工具正常显示。
6. 激活码和使用记录工作区从标签栏下方开始，不再纵向居中。
