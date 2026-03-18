# Go 模块化架构实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有 Go 后端重构为模块化架构，支持插件化管理和公共组件复用

**Architecture:** 模块注册表模式 - 每个模块注册到中央注册表，通过接口暴露能力，事件总线实现模块间异步通信

**Tech Stack:** Go 1.21+, Gin, GORM, Wire, Redis

---

## 文件结构总览

### 新建文件

| 文件路径 | 职责 |
|----------|------|
| `core/registry.go` | 模块注册表，管理所有模块的生命周期 |
| `core/domain/base.go` | 共享领域概念，基础实体定义 |
| `core/event/bus.go` | 事件总线，模块间异步通信 |
| `core/port/port.go` | 模块接口定义占位 |
| `module/category/domain/category.go` | Category 领域模型 |
| `module/category/port/repository.go` | Category Repository 接口 |
| `module/category/port/service.go` | Category Service 接口 |
| `module/category/repository/category_repo.go` | Category Repository 实现 |
| `module/category/service/category_service.go` | Category Service 实现 |
| `module/category/handler/handler.go` | Category HTTP 处理器 |
| `module/category/module.go` | Category 模块注册入口 |
| `module/category/dto.go` | Category DTO 定义 |

### 修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `server/app.go` | 重构为使用模块注册表 |
| `infra/init.go` | 重构为返回 Infra 结构体 |
| `server/middleware/auth.go` | 移动到 `core/middleware/` |
| `server/middleware/cors.go` | 移动到 `core/middleware/` |

---

## Chunk 1: 核心层基础设施

### Task 1: 创建模块注册表

**Files:**
- Create: `core/registry.go`

- [ ] **Step 1: 创建 core 目录**

```bash
mkdir -p core/domain core/port core/event core/middleware
```

- [ ] **Step 2: 编写模块注册表代码**

```go
// core/registry.go
package core

import "github.com/gin-gonic/gin"

// Module 接口 - 所有模块必须实现
type Module interface {
	Name() string
	Routes(api, admin *gin.RouterGroup)
}

// ModuleRegistry 模块注册表
type ModuleRegistry struct {
	modules map[string]Module
}

// NewRegistry 创建新的模块注册表
func NewRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]Module),
	}
}

// Register 注册模块
func (r *ModuleRegistry) Register(m Module) {
	r.modules[m.Name()] = m
}

// Get 获取模块
func (r *ModuleRegistry) Get(name string) Module {
	return r.modules[name]
}

// All 获取所有模块
func (r *ModuleRegistry) All() []Module {
	modules := make([]Module, 0, len(r.modules))
	for _, m := range r.modules {
		modules = append(modules, m)
	}
	return modules
}
```

- [ ] **Step 3: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./core/...
```

- [ ] **Step 4: 提交**

```bash
git add core/registry.go
git commit -m "feat(core): add module registry"
```

---

### Task 2: 创建事件总线

**Files:**
- Create: `core/event/bus.go`

- [ ] **Step 1: 编写事件总线代码**

```go
// core/event/bus.go
package event

import "sync"

// Event 事件接口
type Event interface {
	Name() string
	Payload() any
}

// BasicEvent 基础事件实现
type BasicEvent struct {
	name    string
	payload any
}

func NewEvent(name string, payload any) *BasicEvent {
	return &BasicEvent{name: name, payload: payload}
}

func (e *BasicEvent) Name() string    { return e.name }
func (e *BasicEvent) Payload() any    { return e.payload }

// Handler 事件处理器
type Handler func(Event)

// Bus 事件总线 - 模块间解耦通信
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewBus 创建新的事件总线
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe 订阅事件
func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish 发布事件（异步）
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	handlers := b.handlers[e.Name()]
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(e) // 异步处理
	}
}

// PublishSync 发布事件（同步）
func (b *Bus) PublishSync(e Event) {
	b.mu.RLock()
	handlers := b.handlers[e.Name()]
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./core/...
```

- [ ] **Step 3: 提交**

```bash
git add core/event/bus.go
git commit -m "feat(core): add event bus for module communication"
```

---

### Task 3: 创建共享领域概念

**Files:**
- Create: `core/domain/base.go`
- Create: `core/port/port.go`

- [ ] **Step 1: 编写基础实体定义**

```go
// core/domain/base.go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// Entity 实体接口
type Entity interface {
	GetID() uint
}

// AuditEntity 带审计字段的实体
type AuditEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetID 实现 Entity 接口
func (e *AuditEntity) GetID() uint {
	return e.ID
}
```

- [ ] **Step 2: 创建 port 占位文件**

```go
// core/port/port.go
package port

// 此文件用于存放模块间通信的接口定义
// 各模块的服务接口将在此定义，供其他模块依赖
```

- [ ] **Step 3: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./core/...
```

- [ ] **Step 4: 提交**

```bash
git add core/domain/base.go core/port/port.go
git commit -m "feat(core): add shared domain concepts"
```

---

## Chunk 2: Category 模块迁移

### Task 4: 创建 Category 模块目录结构

**Files:**
- Create directories only

- [ ] **Step 1: 创建模块目录**

```bash
mkdir -p modules/category/{domain,port,repository,service,handler}
```

---

### Task 5: 创建 Category 领域层

**Files:**
- Create: `module/category/domain/category.go`
- Create: `module/category/domain/errors.go`

- [ ] **Step 1: 编写 Category 实体**

```go
// modules/category/domain/category.go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类实体
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null;unique" json:"name"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (*Category) TableName() string {
	return "categories"
}
```

- [ ] **Step 2: 编写错误定义**

```go
// modules/category/domain/errors.go
package domain

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category already exists")
)
```

- [ ] **Step 3: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./modules/category/...
```

- [ ] **Step 4: 提交**

```bash
git add modules/category/domain/
git commit -m "feat(module/category): add domain layer"
```

---

### Task 6: 创建 Category 端口层

**Files:**
- Create: `module/category/port/repository.go`
- Create: `module/category/port/service.go`

- [ ] **Step 1: 编写 Repository 接口**

```go
// modules/category/port/repository.go
package port

import (
	"context"

	"github.com/vidora/vidora-api/module/category/domain"
)

// Repository 分类仓储接口
type Repository interface {
	Create(ctx context.Context, cat *domain.Category) error
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	First(ctx context.Context, id uint) (*domain.Category, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
	ListWithVideoCount(ctx context.Context) ([]domain.Category, map[uint]int64, error)
}
```

- [ ] **Step 2: 编写 Service 接口**

```go
// modules/category/port/service.go
package port

import (
	"context"

	"github.com/vidora/vidora-api/module/category/domain"
)

// Service 分类服务接口（供其他模块调用）
type Service interface {
	Create(ctx context.Context, name string, sortOrder int) (*domain.Category, error)
	Update(ctx context.Context, id uint, name string, sortOrder *int) (*domain.Category, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*domain.Category, error)
	List(ctx context.Context) ([]domain.Category, error)
}
```

- [ ] **Step 3: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./modules/category/...
```

- [ ] **Step 4: 提交**

```bash
git add modules/category/port/
git commit -m "feat(module/category): add port layer interfaces"
```

---

### Task 7: 创建 Category Repository 实现

**Files:**
- Create: `module/category/repository/category_repo.go`

- [ ] **Step 1: 编写 Repository 实现**

```go
// modules/category/repository/category_repo.go
package repository

import (
	"context"

	"github.com/vidora/vidora-api/module/category/domain"
	"github.com/vidora/vidora-api/module/category/port"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

// 确保 Repo 实现了 port.Repository 接口
var _ port.Repository = (*Repo)(nil)

// Repo 分类仓储实现
type Repo struct {
	*dbx.BaseRepo[domain.Category]
	db *gorm.DB
}

// NewRepo 创建分类仓储
func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		BaseRepo: dbx.NewBaseRepo[domain.Category](db),
		db:       db,
	}
}

func (r *Repo) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Category{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *Repo) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Category{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repo) ListWithVideoCount(ctx context.Context) ([]domain.Category, map[uint]int64, error) {
	var categories []domain.Category
	if err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, nil, err
	}

	type countResult struct {
		CategoryID uint
		Count      int64
	}
	var counts []countResult
	r.db.WithContext(ctx).Raw(
		"SELECT category_id, count(*) as count FROM videos WHERE status = 1 AND deleted_at IS NULL GROUP BY category_id",
	).Scan(&counts)

	countMap := make(map[uint]int64)
	for _, c := range counts {
		countMap[c.CategoryID] = c.Count
	}
	return categories, countMap, nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./modules/category/...
```

- [ ] **Step 3: 提交**

```bash
git add modules/category/repository/
git commit -m "feat(module/category): add repository implementation"
```

---

### Task 8: 创建 Category Service 实现

**Files:**
- Create: `module/category/service/category_service.go`

- [ ] **Step 1: 编写 Service 实现**

```go
// modules/category/service/category_service.go
package service

import (
	"context"
	"errors"

	"github.com/vidora/vidora-api/module/category/domain"
	"github.com/vidora/vidora-api/module/category/port"

	"gorm.io/gorm"
)

// 确保 Service 实现了 port.Service 接口
var _ port.Service = (*CategoryService)(nil)

// CategoryService 分类服务实现
type CategoryService struct {
	repo port.Repository
}

// NewCategoryService 创建分类服务
func NewCategoryService(repo port.Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, name string, sortOrder int) (*domain.Category, error) {
	exists, _ := s.repo.ExistsByName(ctx, name)
	if exists {
		return nil, domain.ErrCategoryExists
	}
	cat := &domain.Category{Name: name, SortOrder: sortOrder}
	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Update(ctx context.Context, id uint, name string, sortOrder *int) (*domain.Category, error) {
	updates := make(map[string]interface{})
	if name != "" {
		exists, _ := s.repo.ExistsByName(ctx, name)
		if exists {
			return nil, domain.ErrCategoryExists
		}
		updates["name"] = name
	}
	if sortOrder != nil {
		updates["sort_order"] = *sortOrder
	}
	if len(updates) > 0 {
		s.repo.Update(ctx, id, updates)
	}
	return s.repo.First(ctx, id)
}

func (s *CategoryService) Delete(ctx context.Context, id uint) error {
	exists, _ := s.repo.ExistsByID(ctx, id)
	if !exists {
		return domain.ErrCategoryNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) Get(ctx context.Context, id uint) (*domain.Category, error) {
	cat, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, err
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	categories, _, err := s.repo.ListWithVideoCount(ctx)
	return categories, err
}

// ListWithVideoCount 返回分类列表和视频数量
func (s *CategoryService) ListWithVideoCount(ctx context.Context) ([]domain.Category, map[uint]int64, error) {
	return s.repo.ListWithVideoCount(ctx)
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./modules/category/...
```

- [ ] **Step 3: 提交**

```bash
git add modules/category/service/
git commit -m "feat(module/category): add service implementation"
```

---

### Task 9: 创建 Category Handler 和 DTO

**Files:**
- Create: `module/category/dto.go`
- Create: `module/category/handler/handler.go`

- [ ] **Step 1: 编写 DTO 定义**

```go
// modules/category/dto.go
package category

// CreateCategoryReq 创建分类请求
type CreateCategoryReq struct {
	Name      string `json:"name" binding:"required,max=100"`
	SortOrder int    `json:"sortOrder"`
}

// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	Name      string `json:"name" binding:"max=100"`
	SortOrder *int   `json:"sortOrder"`
}

// CategoryResp 分类响应
type CategoryResp struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	SortOrder  int    `json:"sortOrder"`
	VideoCount int64  `json:"videoCount"`
}
```

- [ ] **Step 2: 编写 Handler 实现**

```go
// modules/category/handler/handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vidora/vidora-api/module/category"
	"github.com/vidora/vidora-api/module/category/service"
	"github.com/vidora/vidora-api/server/helper"
)

// Handler 分类处理器
type Handler struct {
	svc *service.CategoryService
}

// NewHandler 创建分类处理器
func NewHandler(svc *service.CategoryService) *Handler {
	return &Handler{svc: svc}
}

// Create 创建分类
func (h *Handler) Create(c *gin.Context) {
	var req category.CreateCategoryReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	cat, err := h.svc.Create(c, req.Name, req.SortOrder)
	helper.Respond(c, err, cat)
}

// Update 更新分类
func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req category.UpdateCategoryReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	cat, err := h.svc.Update(c, uint(id), req.Name, req.SortOrder)
	helper.Respond(c, err, cat)
}

// Delete 删除分类
func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := h.svc.Delete(c, uint(id))
	helper.Respond(c, err, nil)
}

// List 分类列表
func (h *Handler) List(c *gin.Context) {
	categories, countMap, err := h.svc.ListWithVideoCount(c)
	if err != nil {
		helper.Respond(c, err, nil)
		return
	}

	result := make([]category.CategoryResp, len(categories))
	for i, cat := range categories {
		result[i] = category.CategoryResp{
			ID:         cat.ID,
			Name:       cat.Name,
			SortOrder:  cat.SortOrder,
			VideoCount: countMap[cat.ID],
		}
	}
	helper.Respond(c, nil, result)
}

// Get 分类详情
func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cat, err := h.svc.Get(c, uint(id))
	helper.Respond(c, err, cat)
}
```

- [ ] **Step 3: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./modules/category/...
```

- [ ] **Step 4: 提交**

```bash
git add modules/category/dto.go modules/category/handler/
git commit -m "feat(module/category): add HTTP handlers and DTOs"
```

---

### Task 10: 创建 Category 模块注册入口

**Files:**
- Create: `module/category/module.go`

- [ ] **Step 1: 编写模块注册入口**

```go
// modules/category/modules.go
package category

import (
	"github.com/gin-gonic/gin"
	"github.com/vidora/vidora-api/core"
	"github.com/vidora/vidora-api/module/category/handler"
	"github.com/vidora/vidora-api/module/category/repository"
	"github.com/vidora/vidora-api/module/category/service"

	"gorm.io/gorm"
)

// Module 分类模块
type Module struct {
	Service *service.CategoryService
	Repo    *repository.Repo
	handler *handler.Handler
}

// 确保 Module 实现了 core.Module 接口
var _ core.Module = (*Module)(nil)

// Register 注册分类模块
func Register(registry *core.ModuleRegistry, db *gorm.DB) *Module {
	repo := repository.NewRepo(db)
	svc := service.NewCategoryService(repo)
	h := handler.NewHandler(svc)

	m := &Module{
		Service: svc,
		Repo:    repo,
		handler: h,
	}

	registry.Register(m)
	return m
}

// Name 返回模块名称
func (m *Module) Name() string {
	return "category"
}

// Routes 注册路由
func (m *Module) Routes(api, admin *gin.RouterGroup) {
	// 管理 API
	admin.POST("/categories", m.handler.Create)
	admin.GET("/categories", m.handler.List)
	admin.GET("/categories/:id", m.handler.Get)
	admin.PUT("/categories/:id", m.handler.Update)
	admin.DELETE("/categories/:id", m.handler.Delete)
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add modules/category/modules.go
git commit -m "feat(module/category): add module registration"
```

---

## Chunk 3: 重构应用启动流程

### Task 11: 重构基础设施初始化

**Files:**
- Modify: `infra/init.go`

- [ ] **Step 1: 重构 infra/init.go 返回 Infra 结构体**

将现有全局变量模式改为返回结构体的模式，保持向后兼容：

```go
// infra/init.go
package infra

import (
	"fmt"

	"github.com/lpphub/goweb/pkg/logging"
	"github.com/redis/go-redis/v9"
	"github.com/vidora/vidora-api/module/category/domain"
	"gorm.io/gorm"
)

// Infra 基础设施
type Infra struct {
	Cfg *Config
	DB  *gorm.DB
	RDB *redis.Client
}

// 全局变量（向后兼容）
var (
	Cfg *Config
	DB  *gorm.DB
	RDB *redis.Client
)

// Init 初始化基础设施（旧方式，向后兼容）
func Init() {
	var err error
	// 1.加载配置
	Cfg, err = LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2.配置日志
	logging.Init()

	// 3.初始化数据库和Redis
	DB, err = NewMysqlDB(Cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	RDB, err = NewRedis(Cfg.Redis)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}

	// 4.自动迁移
	autoMigrate()
}

// Bootstrap 初始化基础设施（新方式，返回结构体）
func Bootstrap() (*Infra, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logging.Init()

	db, err := NewMysqlDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	rdb, err := NewRedis(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 自动迁移
	if err := autoMigrateNew(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Infra{
		Cfg: cfg,
		DB:  db,
		RDB: rdb,
	}, nil
}

// ProvideDB 提供 DB（Wire 使用）
func ProvideDB() *gorm.DB {
	return DB
}

// ProvideRDB 提供 Redis（Wire 使用）
func ProvideRDB() *redis.Client {
	return RDB
}

func autoMigrate() {
	err := DB.AutoMigrate(
		// 旧模型路径，后续迁移完成后删除
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate database: %v", err))
	}
}

func autoMigrateNew(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Category{},
		// 其他模块的领域模型将在迁移时添加
	)
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./...
```

- [ ] **Step 3: 提交**

```bash
git add infra/init.go
git commit -m "refactor(infra): add Bootstrap method returning Infra struct"
```

---

### Task 12: 重构应用启动流程

**Files:**
- Modify: `server/app.go`

- [ ] **Step 1: 重构 server/app.go 使用模块注册表**

```go
// server/app.go
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vidora/vidora-api/core"
	"github.com/vidora/vidora-api/core/event"
	"github.com/vidora/vidora-api/infra"
	"github.com/vidora/vidora-api/logic"
	category_module "github.com/vidora/vidora-api/modules/category"
	"github.com/vidora/vidora-api/server/handlers"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/ext/logx"
	"github.com/lpphub/goweb/monitor"
)

type App struct {
	engine   *gin.Engine
	registry *core.ModuleRegistry
	infra    *infra.Infra
	eventBus *event.Bus
	// 旧的服务（向后兼容）
	legacySvc *logic.AppService
}

func New() *App {
	return &App{
		engine:   gin.New(),
		registry: core.NewRegistry(),
		eventBus: event.NewBus(),
	}
}

func (a *App) Run() {
	a.init()
	a.run()
}

func (a *App) init() {
	// 1.初始化基础设施（新方式）
	inf, err := infra.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to bootstrap: %v", err)
	}
	a.infra = inf

	// 2.初始化旧逻辑层（向后兼容）
	logic.Init()
	a.legacySvc = logic.AppSvc

	// 3.注册新模块
	a.registerModules()

	// 4.配置web路由
	a.setupRouter()
}

func (a *App) registerModules() {
	// 注册新模块（按依赖顺序）
	category_module.Register(a.registry, a.infra.DB)
	// 其他模块将在后续迁移中添加
}

func (a *App) setupRouter() {
	r := a.engine

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(logx.GinAccessLog(logx.WithSkipPaths("/metrics", "/health")))

	// pprof and metrics
	monitor.RegisterMetrics(r)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册新模块的路由
	api := r.Group("/api")
	admin := r.Group("/admin")
	for _, m := range a.registry.All() {
		m.Routes(api, admin)
	}

	// 注册旧路由（向后兼容）
	// 注意：新模块的路由会覆盖旧路由
	handlers.RegisterRoutes(r)
}

func (a *App) run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.infra.Cfg.Server.Port),
		Handler: a.engine,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server start failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown completed")
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./...
```

- [ ] **Step 3: 运行测试**

```bash
cd /home/lsk/projects/vidora/vidora-api && go test ./...
```

- [ ] **Step 4: 提交**

```bash
git add server/app.go
git commit -m "refactor(server): integrate module registry into app startup"
```

---

## Chunk 4: 验证与清理

### Task 13: 验证新架构

**Files:**
- None (verification only)

- [ ] **Step 1: 编译整个项目**

```bash
cd /home/lsk/projects/vidora/vidora-api && go build ./...
```

- [ ] **Step 2: 运行所有测试**

```bash
cd /home/lsk/projects/vidora/vidora-api && go test ./...
```

- [ ] **Step 3: 启动服务验证**

```bash
cd /home/lsk/projects/vidora/vidora-api && go run main.go
```

验证分类 API 是否正常工作：
```bash
curl http://localhost:8080/health
curl http://localhost:8080/admin/categories
```

---

### Task 14: 更新文档

**Files:**
- Create: `module/README.md`

- [ ] **Step 1: 编写模块开发指南**

```markdown
# 模块开发指南

## 模块结构

每个模块遵循以下目录结构：

```
module/<name>/
├── domain/           # 领域模型
│   ├── <name>.go     # 实体定义
│   └── errors.go     # 业务错误
├── port/             # 接口定义
│   ├── repository.go # 数据访问接口
│   └── service.go    # 服务接口
├── repository/       # 数据访问实现
├── service/          # 业务逻辑实现
├── handler/          # HTTP 处理器
├── module.go         # 模块注册入口
└── dto.go            # DTO 定义
```

## 添加新模块

1. 创建模块目录结构
2. 实现 `core.Module` 接口
3. 在 `server/app.go` 的 `registerModules()` 中注册

## 模块间通信

- **同步调用**: 通过 `core/port/` 中定义的接口
- **异步通信**: 通过 `core/event.Bus` 发布/订阅事件
```

- [ ] **Step 2: 提交**

```bash
git add modules/README.md
git commit -m "docs: add module development guide"
```

---

## 后续迁移任务（不在本计划范围）

以下模块按相同模式迁移：

1. **User 模块** - 被其他模块依赖，优先迁移
2. **Auth 模块** - 认证相关
3. **Video 模块** - 核心业务模块
4. **Tag 模块**
5. **Transcode 模块**

每个模块迁移完成后，更新 `core/port/` 中的接口定义，并在 `server/app.go` 中注册。

---

## 完成检查清单

- [ ] 核心层（core/）创建完成
- [ ] Category 模块迁移完成
- [ ] 应用启动流程重构完成
- [ ] 编译通过
- [ ] 测试通过
- [ ] 服务启动正常
- [ ] API 功能正常
- [ ] 文档更新完成