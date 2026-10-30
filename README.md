# 万人迷游戏后端（Otome Server）

乙游（乙女向恋爱游戏）后端基础框架，由漫剧 IP 改编项目起步。当前实现账号、
玩家资料、剧情进度三个模块，其余业务模块按领域逐步加入。

## 技术栈

| 层 | 选型 |
|---|---|
| 语言 | Go 1.27 |
| HTTP | Gin |
| 数据库 | MySQL 8.4 + sqlx（显式 SQL） |
| 缓存 | Redis 8（go-redis） |
| 认证 | JWT HS256 + 刷新令牌轮换 |
| 迁移 | golang-migrate（SQL 文件内嵌进二进制） |
| 日志 | log/slog（开发环境文本、生产 JSON） |
| 部署 | Docker Compose / 多阶段 Dockerfile |

消息队列暂不引入，等出现明确异步需求再加。

## 目录结构

```
otome-server/
├─ cmd/
│  ├─ api/                 HTTP 服务入口
│  └─ migrate/             数据库迁移入口
├─ internal/
│  ├─ module/              业务领域模块（handler/service/repository/model）
│  │  ├─ auth/             游客登录、JWT、刷新令牌
│  │  ├─ player/           玩家资料
│  │  └─ story/            剧情进度
│  ├─ platform/            基础设施
│  │  ├─ database/         MySQL 连接池
│  │  ├─ cache/            Redis 客户端
│  │  └─ logger/           slog 初始化
│  ├─ transport/http/      路由、Server 生命周期、中间件
│  └─ shared/              跨模块通用件
│     ├─ apperr/           业务错误
│     ├─ httpx/            统一响应与上下文
│     └─ config/           环境变量配置
├─ api/openapi.yaml        OpenAPI 3.0 接口定义
├─ configs/.env.example    配置样例
├─ deployments/            docker-compose 与 Dockerfile
├─ migrations/             数据库迁移 SQL（内嵌）
├─ scripts/dev.sh          一键启动依赖 + 迁移 + API
├─ test/                   集成 / 端到端测试
├─ docs/architecture.md    架构与设计约束
├─ Makefile
└─ go.mod
```

## 快速开始

前置：Go 1.27+、Docker Desktop。

```bash
# 1. 本地配置（.env 已被 gitignore）
cp configs/.env.example .env

# 2. 启动 MySQL(3307) / Redis(6380)
docker compose -f deployments/docker-compose.yml up -d

# 3. 执行迁移
go run ./cmd/migrate up

# 4. 启动 API（默认 :8090）
go run ./cmd/api
```

或一键执行：`bash scripts/dev.sh`。

验证：

```bash
curl http://127.0.0.1:8090/healthz
curl http://127.0.0.1:8090/readyz
```

> 本机端口说明：`8080` 已被现有 Java 服务占用，`6379` 已被 `huahua-redis`
> 占用，所以本项目使用 **8090 / 3307 / 6380**，不会影响已有项目。

## API

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| GET | `/healthz` | - | 存活探针 |
| GET | `/readyz` | - | 就绪探针（MySQL + Redis） |
| POST | `/api/v1/auth/guest` | - | 游客登录，首次自动注册 |
| POST | `/api/v1/auth/refresh` | - | 刷新令牌（轮换） |
| POST | `/api/v1/auth/logout` | - | 注销刷新令牌 |
| GET | `/api/v1/player/me` | Bearer | 获取玩家资料 |
| PATCH | `/api/v1/player/me` | Bearer | 修改昵称 |
| GET | `/api/v1/story/progress` | Bearer | 获取剧情进度 |
| PUT | `/api/v1/story/progress` | Bearer | 保存剧情进度 |

统一响应：

```json
{"code": 0, "message": "ok", "data": {}, "request_id": "..."}
```

业务码：`0` 成功、`40000` 参数错误、`40100` 未认证、`40300` 无权限、
`40400` 不存在、`40900` 冲突、`50000` 内部错误。

完整定义见 [api/openapi.yaml](api/openapi.yaml)。

### 示例

游客登录：

```bash
curl -s -X POST http://127.0.0.1:8090/api/v1/auth/guest \
  -H "Content-Type: application/json" \
  -d '{"device_id":"dev-device-001","platform":"pc"}'
```

把返回的 `access_token` 填入后续请求：

```bash
TOKEN="<access_token>"

curl -s http://127.0.0.1:8090/api/v1/player/me \
  -H "Authorization: Bearer $TOKEN"

curl -s -X PATCH http://127.0.0.1:8090/api/v1/player/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nickname":"林深"}'

curl -s -X PUT http://127.0.0.1:8090/api/v1/story/progress \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"chapter_id":"ch-01","node_id":"node-12","content_version":"1.0.0","flags":{"affection_lin":3,"choice":"stay"}}'
```

## 配置

所有配置通过环境变量或根目录 `.env` 提供，环境变量优先级更高。

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ENV` | `dev` | `dev` 输出文本日志；`prod` 输出 JSON 且要求强密钥 |
| `HTTP_ADDR` | `:8090` | 监听地址 |
| `MYSQL_DSN` | `...127.0.0.1:3307/otome...` | MySQL DSN，`loc=UTC` |
| `REDIS_ADDR` | `127.0.0.1:6380` | Redis 地址 |
| `JWT_SECRET` | 开发默认值 | 生产环境必须 ≥32 字节 |
| `ACCESS_TOKEN_TTL` | `15m` | 访问令牌有效期 |
| `REFRESH_TOKEN_TTL` | `720h` | 刷新令牌有效期 |
| `LOG_LEVEL` | `info` | `debug/info/warn/error` |
| `SHUTDOWN_TIMEOUT` | `10s` | 优雅退出超时 |

## 数据库迁移

```bash
go run ./cmd/migrate up       # 应用全部迁移
go run ./cmd/migrate down     # 回滚一步
go run ./cmd/migrate version  # 查看当前版本
```

新增迁移：在 `migrations/` 下按 `000002_xxx.up.sql` / `000002_xxx.down.sql`
命名，`embed.go` 会自动打包。

当前表：`users`、`players`、`refresh_tokens`、`story_progress`。设计说明见
[docs/architecture.md](docs/architecture.md)。

## 开发

```bash
go test ./...        # 单元测试（无需 MySQL/Redis）
go vet ./...
gofmt -l .           # 检查格式
go build ./...
```

Windows 默认没有 `make`，可直接使用上面的 `go` 命令；Linux/macOS 可使用
`Makefile` 中的 `make run`、`make migrate-up`、`make compose-up` 等目标。

## 后续规划

1. economy：钱包与双录流水，统一货币入口；
2. content：剧情内容包 manifest、版本与灰度、OSS/CDN；
3. gacha：服务端 RNG、保底、卡池版本化与审计；
4. shop / payment：订单、回调幂等、对账；
5. ops / admin：邮件、公告、兑换码、补偿与运营后台。

注意事项与设计约束集中在 [docs/architecture.md](docs/architecture.md)。
