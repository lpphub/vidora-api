# Go 后端插件化架构设计

**日期**: 2026-03-10
**状态**: 已批准

## 目标

设计一个模块化的 Go 后端架构，实现：
- 模块独立开发与测试
- 清晰的模块边界和接口
- 公共组件复用
- 为未来微服务拆分预留空间

## 约束

- **团队规模**: 1-3 人小团队
- **插件模式**: 模块化单体 + 可拆分（无明确拆分计划）
- **技术栈**: Gin + GORM + Wire（保持现有）

## 架构方案

采用 **模块注册表模式**：每个模块注册到中央注册表，通过接口暴露能力。

## 目录结构

```
vidora-api/
├── cmd/                        # 应用入口
│   └── server/main.go
├── core/                       # 核心层 (公共组件)
│   ├── domain/                 # 共享领域概念 (基础实体、值对象)
│   ├── port/                   # 模块接口定义 (模块间通信契约)
│   ├── event/                  # 事件总线 (模块间解耦通信)
│   └── middleware/             # HTTP中间件 (认证、日志、限流等)
├── infra/                      # 基础设施实现
│   ├── persistence/            # 数据库连接、基础 Repository
│   ├── cache/                  # Redis 封装
│   ├── mq/                     # 消息队列封装 (未来)
│   ├── storage/                # 文件存储封装
│   └── config/                 # 配置加载
├── module/                     # 业务模块
│   ├── user/
│   ├── video/
│   ├── category/
│   └── ...                     # 其他模块
├── server/                     # HTTP服务
│   ├── app.go                  # 应用启动
│   ├── router.go               # 路由注册
│   └── response/               # 统一响应
├── pkg/                        # 通用工具
│   ├── errors/
│   ├── utils/
│   └── validator/
└── api/                        # API定义 (可选)
    └── openapi/
```

## 核心组件

### 1. 模块注册表 (`core/registry.go`)

```go
package core

// Module 接口 - 所有模块必须实现
type Module interface {
    Name() string
    Routes(r *gin.RouterGroup)
}

// ModuleRegistry 模块注册表
type ModuleRegistry struct {
    modules map[string]Module
}

func NewRegistry() *ModuleRegistry {
    return &ModuleRegistry{
        modules: make(map[string]Module),
    }
}

func (r *ModuleRegistry) Register(m Module) {
    r.modules[m.Name()] = m
}

func (r *ModuleRegistry) Get(name string) Module {
    return r.modules[name]
}

func (r *ModuleRegistry) All() []Module {
    modules := make([]Module, 0, len(r.modules))
    for _, m := range r.modules {
        modules = append(modules, m)
    }
    return modules
}
```

### 2. 模块接口定义 (`core/port/`)

```go
// core/port/user_port.go
package port

// UserService 用户服务接口 - 供其他模块调用
type UserService interface {
    GetByID(ctx context.Context, id uint) (*user.User, error)
    GetByEmail(ctx context.Context, email string) (*user.User, error)
}

// VideoService 视频服务接口
type VideoService interface {
    GetByID(ctx context.Context, id uint) (*video.Video, error)
    Publish(ctx context.Context, id uint) error
}
```

### 3. 事件总线 (`core/event/bus.go`)

```go
package event

// Event 事件接口
type Event interface {
    Name() string
    Payload() any
}

// Handler 事件处理器
type Handler func(Event)

// Bus 事件总线 - 模块间解耦通信
type Bus struct {
    handlers map[string][]Handler
}

func NewBus() *Bus {
    return &Bus{
        handlers: make(map[string][]Handler),
    }
}

// Subscribe 订阅事件
func (b *Bus) Subscribe(eventName string, h Handler) {
    b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish 发布事件
func (b *Bus) Publish(e Event) {
    for _, h := range b.handlers[e.Name()] {
        go h(e) // 异步处理
    }
}
```

## 模块结构

每个模块采用一致的内部结构：

```
module/video/
├── domain/                     # 领域层
│   ├── video.go                # 实体定义
│   ├── events.go               # 领域事件
│   └── errors.go               # 业务错误
├── port/                       # 端口层 (接口定义)
│   ├── repository.go           # 数据访问接口
│   └── service.go              # 服务接口 (供其他模块调用)
├── service/                    # 应用服务层
│   └── video_service.go        # 业务逻辑实现
├── repository/                 # 适配器层
│   └── video_repo.go           # 数据访问实现
├── handler/                    # HTTP处理器
│   ├── admin.go                # 管理端API
│   └── api.go                  # 公开API
├── module.go                   # 模块注册入口
└── dto.go                      # DTO定义
```

### 模块注册入口示例

```go
// module/video/module.go
package video

type Module struct {
    Service    *service.VideoService
    Repository port.Repository
}

// Register 注册模块到系统
func Register(registry *core.ModuleRegistry, db *gorm.DB, rdb *redis.Client, eventBus *event.Bus) *Module {
    repo := repository.NewVideoRepo(db)
    svc := service.NewVideoService(repo, rdb, eventBus)

    module := &Module{
        Service:    svc,
        Repository: repo,
    }

    registry.Register(module)
    return module
}

func (m *Module) Name() string {
    return "video"
}

func (m *Module) Routes(api, admin *gin.RouterGroup) {
    // 公开 API
    api.GET("/videos", m.handler.List)
    api.GET("/videos/:id", m.handler.Get)

    // 管理 API
    admin.POST("/videos", m.handler.Create)
    admin.PUT("/videos/:id", m.handler.Update)
    admin.DELETE("/videos/:id", m.handler.Delete)
}
```

## 应用启动流程

```go
// server/app.go
package server

type App struct {
    engine   *gin.Engine
    registry *core.ModuleRegistry
    infra    *infra.Infra
    eventBus *event.Bus
}

func New() *App {
    return &App{
        engine:   gin.New(),
        registry: core.NewRegistry(),
        eventBus: event.NewBus(),
    }
}

func (a *App) Run() {
    // 1. 初始化基础设施
    infra, err := infra.Bootstrap()
    if err != nil {
        log.Fatal(err)
    }
    a.infra = infra

    // 2. 注册所有模块
    a.registerModules()

    // 3. 配置路由
    a.setupRouter()

    // 4. 启动服务
    a.start()
}

func (a *App) registerModules() {
    // 按依赖顺序注册模块
    user.Register(a.registry, a.infra.DB, a.infra.RDB, a.eventBus)
    auth.Register(a.registry, a.infra.DB, a.eventBus)
    category.Register(a.registry, a.infra.DB)
    video.Register(a.registry, a.infra.DB, a.infra.RDB, a.eventBus)
}

func (a *App) setupRouter() {
    r := a.engine

    // 全局中间件
    r.Use(gin.Recovery())
    r.Use(middleware.Logger())

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 注册所有模块的路由
    api := r.Group("/api")
    admin := r.Group("/admin", middleware.JWTAuth())

    for _, m := range a.registry.All() {
        m.Routes(api, admin)
    }
}
```

## 模块间通信

### 方式1：通过接口调用（同步）

```go
// module/video/service/video_service.go
type VideoService struct {
    repo       port.Repository
    userSvc    port.UserService  // 注入用户服务接口
}

func (s *VideoService) GetWithAuthor(ctx context.Context, videoID uint) (*VideoDetail, error) {
    video, err := s.repo.GetByID(ctx, videoID)
    if err != nil {
        return nil, err
    }

    // 通过接口调用其他模块
    author, err := s.userSvc.GetByID(ctx, video.AuthorID)
    if err != nil {
        return nil, err
    }

    return &VideoDetail{
        Video:  video,
        Author: author,
    }, nil
}
```

### 方式2：通过事件总线（异步）

```go
// module/video/service/video_service.go
func (s *VideoService) Publish(ctx context.Context, id uint) error {
    // 业务逻辑...

    // 发布事件
    s.eventBus.Publish(event.Event{
        Name: "video.published",
        Payload: VideoPublishedEvent{VideoID: id},
    })
    return nil
}

// module/notification/module.go
func (m *Module) Init(eventBus *event.Bus) {
    eventBus.Subscribe("video.published", func(e event.Event) {
        // 发送通知...
    })
}
```

## 迁移策略

### 阶段1：创建核心层（不影响现有代码）

```bash
mkdir -p core/{domain,port,event,middleware}
```

实现：
- `core/registry.go` - 模块注册表
- `core/event/bus.go` - 事件总线
- `core/port/*.go` - 模块接口定义

### 阶段2：迁移一个模块作为示例

选择最简单的模块（如 category）：

```bash
mkdir -p module/category/{domain,port,service,repository,handler}
```

迁移步骤：
1. `domain/` - 移动实体定义
2. `port/` - 定义 Repository 和 Service 接口
3. `repository/` - 实现数据访问
4. `service/` - 实现业务逻辑
5. `handler/` - 实现 HTTP 处理器
6. `module.go` - 模块注册入口

### 阶段3：更新启动流程

修改 `server/app.go` 使用模块注册表。

### 阶段4：迁移其他模块

按依赖关系依次迁移：

| 优先级 | 模块 | 理由 |
|--------|------|------|
| 1 | category | 结构最简单，适合作为示例 |
| 2 | user | 被其他模块依赖，需要先迁移 |
| 3 | video | 核心业务模块 |
| 4 | 其他模块 | 按依赖关系依次迁移 |

### 阶段5：清理旧代码

删除 `logic/` 目录中已迁移的代码。

## 公共组件清单

| 目录 | 职责 | 内容 |
|------|------|------|
| `core/domain` | 共享领域概念 | `Entity`, `AuditEntity`, 基础值对象 |
| `core/port` | 模块接口 | `UserService`, `VideoService` 等 |
| `core/event` | 事件总线 | `Bus`, `Event`, `Handler` |
| `core/middleware` | HTTP中间件 | JWT认证、日志、限流、CORS |
| `pkg/errors` | 错误处理 | 统一错误码、错误包装 |
| `pkg/utils` | 工具函数 | 字符串、时间、加密等 |
| `pkg/validator` | 验证器 | 请求验证、自定义规则 |

## 设计决策记录

1. **为什么不使用 Go 原生插件？**
   - Go 插件 (.so) 有平台限制和稳定性问题
   - 小团队不需要运行时动态加载
   - 编译时模块化已足够满足需求

2. **为什么不用六边形架构？**
   - 六边形架构代码量较多
   - 小团队更看重开发效率
   - 模块注册表模式已提供足够的解耦

3. **为什么保留 Wire？**
   - 团队已熟悉 Wire
   - 编译时依赖注入更安全
   - 模块内使用 Wire 初始化依赖

## 未来扩展

如果需要拆分为微服务：

1. 将模块接口 (`core/port/`) 转换为 gRPC 定义
2. 模块间调用从本地接口改为 gRPC 调用
3. 事件总线替换为消息队列（Kafka/RabbitMQ）
4. 各模块独立部署