# Vidora API 架构重构设计

## 背景

当前项目架构存在以下问题：
- `biz/` 目录定位模糊，只有 2 个文件
- `biz/port/port.go` 导入了具体模块的 domain，违反依赖倒置原则
- `server/app.go` 手动传递模块依赖，耦合度高
- `pkg/core/` 放置核心注册逻辑，位置不合适
- 模块内部嵌套过深（domain/handler/repository/service 子目录）

## 目标

- 结构清晰简洁，适合中小型个人项目
- 解决循环依赖问题
- 支持模块化开发，新增功能减少影响范围
- 使用 Wire 管理依赖注入

## 架构设计

### 目录结构

```
vidora-api/
├── cmd/
│   └── server/
│       ├── main.go           # 入口
│       └── wire.go           # Wire 注入定义
├── contract/                  # 模块间接口契约
│   ├── user.go
│   ├── video.go
│   └── category.go
├── infra/                     # 基础设施
│   ├── config.go
│   ├── database.go
│   ├── redis.go
│   └── wire.go
├── module/                    # 业务模块
│   ├── user/
│   │   ├── entity.go         # 实体定义
│   │   ├── dto.go            # 请求/响应结构
│   │   ├── errors.go         # 模块错误
│   │   ├── repository.go     # 数据访问
│   │   ├── service.go        # 业务逻辑
│   │   ├── handler.go        # HTTP 处理
│   │   ├── module.go         # 模块组装
│   │   └── wire.go           # ProviderSet
│   ├── video/
│   │   └── ...
│   └── wire.go               # 合并所有模块
├── server/
│   ├── app.go
│   ├── helper/
│   │   └── helper.go         # HTTP 响应辅助方法
│   ├── middleware/
│   │   ├── auth.go
│   │   └── cors.go
│   └── wire.go
├── pkg/                       # 工具包
│   ├── errs/
│   ├── strutils/
│   └── event/
├── config/                    # 配置文件
├── go.mod
└── Makefile
```

### 依赖规则

```
cmd/server    → server, infra, module
server        → infra, module, pkg
module/*      → contract, infra, pkg
contract      → (无依赖)
infra         → 第三方库 (gorm, redis, viper 等)
pkg           → (无内部依赖)
```

**核心原则：** 模块之间通过 `contract` 接口通信，零直接依赖。

### 各层职责

| 目录 | 职责 |
|------|------|
| `cmd/server` | 应用入口，Wire 初始化 |
| `contract` | 模块间通信接口定义，纯接口 + DTO |
| `infra` | 基础设施，全局变量 (DB, RDB, Cfg) |
| `module` | 业务模块实现，每个模块自包含 |
| `server` | HTTP 服务配置、中间件、辅助方法 |
| `pkg` | 可复用工具包，无业务逻辑 |

### Wire 依赖注入

#### infra/wire.go

```go
package infra

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewConfig,
    NewDB,
    NewRedis,
)
```

#### module/user/wire.go

```go
package user

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewRepository,
    NewService,
    NewHandler,
    NewModule,
)
```

#### module/wire.go

```go
package module

import (
    "github.com/google/wire"
    "vidora-api/module/user"
    "vidora-api/module/video"
)

var ProviderSet = wire.NewSet(
    user.ProviderSet,
    video.ProviderSet,
)
```

#### cmd/server/wire.go

```go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "vidora-api/infra"
    "vidora-api/module"
    "vidora-api/server"
)

func InitializeApp() (*server.App, func(), error) {
    wire.Build(
        infra.ProviderSet,
        module.ProviderSet,
        server.ProviderSet,
    )
    return nil, nil, nil
}
```

### 模块结构

每个模块包含以下文件：

| 文件 | 职责 |
|------|------|
| `entity.go` | 数据库实体，包含基础字段如 AuditEntity |
| `dto.go` | 请求/响应数据结构 |
| `errors.go` | 模块内定义的错误 |
| `repository.go` | 数据访问层，依赖 `*gorm.DB` |
| `service.go` | 业务逻辑，实现 `contract` 接口 |
| `handler.go` | HTTP 处理器，注册路由 |
| `module.go` | 组装所有组件，统一对外暴露 |
| `wire.go` | Wire ProviderSet |

### 模块间通信

模块 A 需要调用模块 B 时：

1. 在 `contract/` 定义接口：
```go
// contract/user.go
package contract

type UserService interface {
    Get(ctx context.Context, id uint) (*UserDTO, error)
}

type UserDTO struct {
    ID    uint
    Email string
    Name  string
}
```

2. 模块 B 的 Service 实现接口：
```go
// modules/user/service.go
var _ contract.UserService = (*Service)(nil)

func (s *Service) Get(ctx context.Context, id uint) (*contract.UserDTO, error) {
    // ...
}
```

3. 模块 A 通过接口注入依赖：
```go
// modules/video/modules.go
func NewModule(
    db *gorm.DB,
    userSvc contract.UserService,  // 接口类型
) *Module {
    // ...
}
```

4. Wire 自动处理依赖注入

### 文件迁移映射

| 原路径 | 新路径 |
|--------|--------|
| `biz/port/port.go` | `contract/*.go` |
| `biz/domain/base.go` | 各模块 `entity.go` 内部定义 |
| `pkg/core/registry.go` | 删除（由 Wire 替代） |
| `infra/init.go` | 拆分为 `infra/*.go` + `wire.go` |
| `module/*/domain/*` | `module/*/entity.go`, `dto.go`, `errors.go` |
| `module/*/repository/*` | `module/*/repository.go` |
| `module/*/service/*` | `module/*/service.go` |
| `module/*/handler/*` | `module/*/handler.go` |

### 新增模块流程

1. 创建 `module/xxx/` 目录
2. 实现各层文件 (entity/dto/errors/repository/service/handler/module/wire)
3. 如需被其他模块调用，在 `contract/` 定义接口
4. 在 `module/wire.go` 添加 ProviderSet
5. 运行 `wire ./cmd/server/` 生成代码

## 风险与对策

| 风险 | 对策 |
|------|------|
| 迁移工作量大 | 分批迁移，先迁移一个模块验证流程 |
| Wire 学习成本 | 文档已提供示例，编译时检查保证类型安全 |
| 基础实体重复定义 | 按模块定义，避免过度抽象 |