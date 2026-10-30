# otome-server

乙游后端，Go 模块化单体。完整说明见 README.md 与 docs/architecture.md。

## 常用命令

- 启动依赖：`docker compose -f deployments/docker-compose.yml up -d`
- 数据库迁移：`go run ./cmd/migrate up`
- 启动 API：`go run ./cmd/api`（:8090）
- 单元测试：`go test ./...`
- 格式检查：`gofmt -l .`

## 目录约定

- `internal/module/<domain>/`：业务模块，包含 handler / service / repository / model；
  模块之间不要互相导入。
- `internal/platform/`：MySQL、Redis、日志等基础设施。
- `internal/transport/http/`：路由、中间件、Server 生命周期。
- `internal/shared/`：config、apperr、httpx 等通用件。
- `cmd/` 只做组装，不写业务逻辑。

## 约束

- 本机端口：API `8090`，MySQL `3307`，Redis `6380`（避开已有 Java 服务与 huahua 容器）。
- 暂不引入消息队列；有明确异步需求时再加。
- 禁止 N+1、循环 SQL、内存分页；写操作必须支持幂等。
- 时间统一使用 UTC。
- 新增模块：先建 `internal/module/<name>/`，再在 `internal/transport/http/router.go` 注册路由。
