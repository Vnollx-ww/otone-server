# 架构说明

## 分层

```
cmd/api          HTTP 服务入口：加载配置、连接 MySQL/Redis、启动 HTTP Server
cmd/migrate      数据库迁移入口

internal/
  module/        业务领域模块，每个模块自包含 handler / service / repository / model
    auth/        游客账号、JWT、刷新令牌轮换
    player/      玩家资料
    story/       剧情进度

  platform/      与技术栈绑定的基础设施
    database/    MySQL 连接池
    cache/       Redis 客户端
    logger/      slog 日志

  transport/http/  HTTP 适配层
    router.go      路由组装、健康检查
    server.go      HTTP Server 生命周期与优雅退出
    middleware/    请求 ID、访问日志、panic 恢复、JWT 鉴权

  shared/        跨模块通用件
    apperr/      业务错误（带业务码与 HTTP 状态码）
    httpx/       统一响应封装与上下文取值
    config/      环境变量配置
```

依赖方向：`cmd` → `transport` / `platform` / `module`；`transport` → `module`；
`module` → `platform` / `shared`。业务模块之间不要互相导入，跨模块协作通过
service 接口或后续的领域事件完成。

## 请求链路

1. `RequestID` 读取或生成 `X-Request-ID`；
2. `AccessLog` 记录方法、路径、状态码、耗时、用户 ID；
3. `Recovery` 捕获 panic，返回统一错误结构；
4. 鉴权中间件解析 Bearer JWT，把 `user_id` 写入上下文；
5. handler 绑定参数 → service 校验/业务处理 → repository 执行 SQL；
6. handler 通过 `httpx.OK` / `httpx.Fail` 输出统一结构。

统一响应：

```json
{"code": 0, "message": "ok", "data": {}, "request_id": "..."}
```

业务码约定：`0` 成功，`40000` 参数错误，`40100` 未认证，`40300` 无权限，
`40400` 不存在，`40900` 冲突，`50000` 内部错误。

## 数据设计

- `users`：账号主体，`(device_id, platform)` 唯一，支持游客自动注册；
- `players`：玩家资料，与 `users` 一对一；
- `refresh_tokens`：只存 SHA-256 哈希，刷新时轮换并撤销旧令牌；
- `story_progress`：按 `(user_id, chapter_id)` 保存剧情快照，`flags` 为 JSON，
  用于好感度、分支标记等；更新时间由数据库统一写入。

所有时间使用 UTC（DSN `loc=UTC`，容器 `default-time-zone=+00:00`）。

## 设计约束

- 禁止循环 SQL / N+1：列表查询使用 `IN` 批量查询或 `JOIN`；
- 列表接口必须数据库分页（`LIMIT/OFFSET` 或游标）；剧情章节是有界集合，
  当前 `LIMIT 200` 只是安全上限；
- 货币、抽卡、支付等写操作必须具备幂等键，服务端为权威数据源；
- 错误只通过 `apperr` 暴露稳定业务码，内部错误详情只写日志；
- 业务模块不依赖 HTTP：handler 负责协议转换，service 不导入 gin。

## 后续模块

1. **economy**：钱包 + 双录流水（ledger），所有货币变动统一入口；
2. **content**：剧情内容包 manifest、版本与灰度，资源走 OSS/CDN；
3. **gacha**：服务端 RNG、保底计数、卡池配置版本化、审计日志；
4. **shop / payment**：订单、回调幂等、收据校验、对账；
5. **ops**：邮件、公告、兑换码、补偿、热配置；
6. **admin**：运营后台 API（RBAC），前端独立仓库或 `web/` 子目录。

消息队列暂不引入；当出现明确的异步需求（发奖、埋点、推送）时再接入。
