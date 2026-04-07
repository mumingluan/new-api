# zch-api 提交记录备份

## 2026-03-24: Redis Sentinel 故障转移重试逻辑

### 修改内容
- 修改 `middleware/rate-limit.go`，为限流函数添加 Redis 重试逻辑
- 修改函数：`redisRateLimiter`、`userRedisRateLimiter`

### 问题背景
- `go-redis` 的 `FailoverClient` 本身支持热切换，但在 Sentinel 切换 master 的短暂时间（几百毫秒）内，Redis 连接会失败
- 原代码在 Redis 操作失败时直接返回 500 错误并拒绝请求，导致 `rate_limit_check_failed`
- 用户在 Sentinel 故障转移期间会遇到限流检查失败，需要重启服务才能恢复

### 解决方案
**添加重试机制：**
- Redis 操作失败时自动重试最多 3 次
- 每次重试间隔 100ms
- 重试逻辑应用于 `LLen` 等关键操作

**Fail-open 策略：**
- 如果 3 次重试后仍失败，采用 fail-open 策略（允许请求通过）
- 避免 Redis 短暂不可用时完全阻断服务
- 在日志中记录错误："Redis rate limit error after retries"

### 代码示例
```go
// Retry logic for Sentinel failover
var listLength int64
var err error
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    listLength, err = rdb.LLen(ctx, key).Result()
    if err == nil {
        break
    }
    if i < maxRetries-1 {
        time.Sleep(time.Millisecond * 100)
    }
}

if err != nil {
    fmt.Println("Redis rate limit error after retries:", err.Error())
    c.Next() // Fail-open: allow request
    return
}
```

### 效果
- Sentinel 切换 master 时，大部分请求会在重试中成功（100-200ms 延迟）
- 少数请求会被 fail-open 放行（避免服务中断）
- 无需重启服务即可平滑处理 Redis 故障转移

### 部署状态
- ✓ NekoMetal 已部署（2026-03-24 06:48）
- ⏳ Metal 待部署

---

## 2026-03-23: Redis Sentinel 支持

### 修改内容
- 修改 `common/redis.go`，添加 Redis Sentinel 高可用支持
- 新增环境变量：
  - `REDIS_SENTINEL_ADDRS`: Sentinel 地址列表（逗号分隔）
  - `REDIS_SENTINEL_MASTER_NAME`: Master 名称
  - `REDIS_PASSWORD`: Redis 密码
  - `REDIS_DB`: Redis 数据库编号
- 保持向后兼容：如果未设置 Sentinel 变量，仍使用原有的 `REDIS_CONN_STRING` 单机模式

### 实现细节
- 使用 `redis.NewFailoverClient` 连接 Sentinel 集群
- 添加辅助函数：`parseCommaSeparated`、`splitString`、`trimSpace`、`GetEnvOrDefaultInt`
- Sentinel 模式下自动追踪当前 master，无需手动切换

### 部署架构
- Metal (10.200.0.5): Redis master + Sentinel
- NekoMetal (10.200.0.3): Redis replica + Sentinel
- mumlECS (10.200.0.1): Sentinel only
- Quorum: 2/3（需要至少 2 个 Sentinel 同意才能故障转移）
- Sentinel 超时时间：5 分钟（300000ms）

### 配置示例
```bash
# Sentinel 模式
REDIS_SENTINEL_ADDRS=10.200.0.5:26379,10.200.0.3:26379,10.200.0.1:26379
REDIS_SENTINEL_MASTER_NAME=mymaster
REDIS_PASSWORD=SrF503117db
REDIS_DB=0

# 单机模式（向后兼容）
REDIS_CONN_STRING=redis://:password@localhost:6379/0
```

### 故障转移行为
- Metal Redis 宕机或无响应超过 5 分钟 → Sentinel 自动提升 NekoMetal 为新 master
- Metal 恢复后 → 自动变成 replica，跟随当前 master
- 客户端自动追踪 master 变化，无需重启服务（配合重试逻辑）

### 测试状态
- ✓ NekoMetal new-api 成功连接 Sentinel
- ✓ Metal new-api 成功连接 Sentinel
- ✓ Sentinel 集群正常监控 master
- ✓ 限流数据在主从间实时同步
- ✓ 手动故障转移测试通过（Metal ↔ NekoMetal）
