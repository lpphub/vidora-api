# Vidora API

视频管理平台后端服务。

## 技术栈

- **Go 1.25** + **Gin** + **GORM**
- **MySQL** + **Redis**
- **JWT** 认证
- **lpphub/goweb** 工具库

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────────────┐
│                   main.go                        │
│                  (入口启动)                       │
└──────────────────────┬──────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────────┐
│                   server/                        │
│             (HTTP、路由、中间件)                   │
└──────────────────────┬───────────────────────────┘
                       │
           ┌───────────┴───────────────────────────┐
           │                                       │
           ▼                                       ▼
┌───────────────────────────────┐      ┌───────────────────────┐
│           module/             │      │        infra/         │
│                               │      │  DB、Redis、Config    │
│  handler  →  service  →  repo │──────┤──注入                 │
│                               │      │                       │
└───────────────┬───────────────┘      └───────────────────────┘
                │
                │ 实现
                ▼
┌──────────────────────┐
│      contract/       │
│   模块间接口契约      │
│  解耦、避免循环依赖    │
└──────────────────────┘
```

### 项目结构

```
vidora-api/
├── config/          # 配置文件
├── contract/        # 跨模块接口契约
├── infra/           # 基础设施（DB、Redis、JWT）
├── module/          # 业务模块
├── server/          # HTTP 服务（路由、中间件）
├── shared/          # 共享工具（错误、常量）
├── main.go          # 程序入口
└── Makefile
```

**module/** - 业务模块，核心代码所在：

```
module/video/
├── model.go        # 数据库实体、领域行为（依赖：GORM）
├── dto.go          # HTTP 请求/响应结构
├── repository.go   # 数据库 CRUD（依赖：GORM、Model）
├── service.go      # 业务逻辑（依赖：Repository、contract）
├── handler.go      # HTTP 处理、路由注册（依赖：Service、DTO）
└── init.go         # 依赖注入、模块组装
```

**调用流程：**

```
HTTP → Handler → Service → Repository → DB
                     │
                     └→ contract 接口 → 其他模块
```

**contract/** - 跨模块接口契约，用于解耦和解决循环依赖：

```go
// contracts/user.go
type UserBiz interface {
    Create(ctx context.Context, email, password string) (*UserDTO, error)
}
```

- 模块间调用通过接口，不直接依赖具体实现
- 避免循环引用：`video` 调用 `contract.UserBiz`，`user` 实现接口

**Repository 基类：**

继承 `dbx.BaseRepo[T]` 自动获得基础 CRUD：

```go
type Repository struct {
    *dbx.BaseRepo[Video]  // Create、First、Update、Delete 等
}

func NewRepository(db *gorm.DB) *Repository {
    return &Repository{BaseRepo: dbx.NewBaseRepo[Video](db)}
}
```

## 快速开始

### 环境要求

Go 1.25+、MySQL 8.0+、Redis 6.0+

### 配置

编辑 `config/config.yml`：

```yaml
database:
  host: localhost
  port: 3306
  dbname: vidora
  user: root
  password: your_password

redis:
  host: localhost
  port: 6379

jwt:
  secret: your_jwt_secret

server:
  port: 8080
```

### 运行

```bash
go mod tidy
make dev
```

### 部署

```bash
make build
docker build -t vidora-api .
```

## 特性

- 模块化分层架构
- 通过接口解耦，避免循环引用
- JWT 认证
- 优雅关闭
- Prometheus 监控
