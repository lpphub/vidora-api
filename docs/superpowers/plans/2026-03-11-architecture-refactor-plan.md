# Vidora API 架构重构实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构项目架构为清晰的模块化结构，使用 Wire 管理依赖注入，解决循环依赖问题。

**Architecture:** 扁平化模块结构，模块间通过 contract 接口通信，Wire 编译时注入依赖，基础设施保持全局变量。

**Tech Stack:** Go 1.25, Gin, GORM, Wire, Redis

---

## Chunk 1: 基础结构搭建

### Task 1: 创建目录结构

**Files:**
- Create: `cmd/server/` 目录
- Create: `contract/` 目录
- Delete: `biz/` 目录
- Delete: `pkg/core/` 目录

- [ ] **Step 1: 创建新目录结构**

```bash
mkdir -p cmd/server
mkdir -p contract
```

- [ ] **Step 2: 移动 main.go 到 cmd/server**

```bash
mv main.go cmd/server/main.go
```

- [ ] **Step 3: 删除旧目录**

```bash
rm -rf biz/
rm -rf pkg/core/
```

- [ ] **Step 4: 验证目录结构**

Run: `ls -la`
Expected: 看到 `cmd/`, `contract/`, `infra/`, `module/`, `server/`, `pkg/`

---

### Task 2: 创建 contract 包 - 接口定义

**Files:**
- Create: `contract/user.go`
- Create: `contract/category.go`
- Create: `contract/tag.go`
- Create: `contract/video.go`

- [ ] **Step 1: 创建 contract/user.go**

```go
// contract/user.go
package contract

import "context"

// UserService 用户服务接口（供其他模块调用）
type UserService interface {
	Create(ctx context.Context, email, password string) (*UserDTO, error)
	Get(ctx context.Context, userID uint) (*UserDTO, error)
	GetByEmail(ctx context.Context, email string) (*UserDTO, error)
	GetByIDs(ctx context.Context, ids []uint) ([]UserDTO, error)
	ValidateLogin(ctx context.Context, email, password string) (*UserDTO, error)
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID       uint
	Email    string
	Name     string
	Password string
}
```

- [ ] **Step 2: 创建 contract/category.go**

```go
// contract/category.go
package contract

import "context"

// CategoryService 分类服务接口（供 Video 等模块调用）
type CategoryService interface {
	Get(ctx context.Context, id uint) (*CategoryDTO, error)
}

// CategoryDTO 分类数据传输对象
type CategoryDTO struct {
	ID          uint
	Name        string
	Description string
}
```

- [ ] **Step 3: 创建 contract/tag.go**

```go
// contract/tag.go
package contract

import "context"

// TagService 标签服务接口（供 Video 等模块调用）
type TagService interface {
	ValidateTagIDs(ctx context.Context, ids []uint) error
}

// TagRepository 标签仓储接口（供 Video 模块调用）
type TagRepository interface {
	SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error
}

// TagDTO 标签数据传输对象
type TagDTO struct {
	ID   uint
	Name string
}
```

- [ ] **Step 4: 创建 contract/video.go**

```go
// contract/video.go
package contract

import "context"

// VideoService 视频服务接口（供外部调用）
type VideoService interface {
	Create(ctx context.Context, req CreateVideoReq) (*VideoDTO, error)
	Update(ctx context.Context, id uint, req UpdateVideoReq) (*VideoDTO, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*VideoDTO, error)
	List(ctx context.Context, req VideoListReq) (*VideoListDTO, error)
}

// CreateVideoReq 创建视频请求
type CreateVideoReq struct {
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	TagIDs      []uint
}

// UpdateVideoReq 更新视频请求
type UpdateVideoReq struct {
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	Status      *int8
	TagIDs      []uint
}

// VideoListReq 视频列表请求
type VideoListReq struct {
	CategoryID uint
	Status     *int8
	Keyword    string
	Page       int
	PageSize   int
}

// VideoDTO 视频数据传输对象
type VideoDTO struct {
	ID          uint
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	Status      int8
	PlayCount   int
}

// VideoListDTO 视频列表响应
type VideoListDTO struct {
	Total int64
	List  []VideoDTO
}
```

- [ ] **Step 5: 验证编译**

Run: `go build ./contract/...`
Expected: 无错误

---

### Task 3: 重构 infra 包

**Files:**
- Modify: `infra/init.go` → 拆分为多个文件
- Create: `infra/config.go`
- Create: `infra/database.go`
- Create: `infra/redis.go`
- Create: `infra/wire.go`
- Delete: `infra/dbs.go`

- [ ] **Step 1: 读取现有 infra 代码**

Run: `cat infra/init.go infra/config.go infra/dbs.go`

- [ ] **Step 2: 创建 infra/config.go**

```go
// infra/config.go
package infra

import (
	"github.com/lpphub/goweb/pkg/config"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

var Cfg *Config

// NewConfig 加载配置
func NewConfig() (*Config, error) {
	Cfg = &Config{
		Server: ServerConfig{
			Port: config.GetInt("server.port", 8080),
		},
		Database: DatabaseConfig{
			Host:     config.GetString("database.host", "localhost"),
			Port:     config.GetInt("database.port", 3306),
			User:     config.GetString("database.user", "root"),
			Password: config.GetString("database.password", ""),
			DBName:   config.GetString("database.name", "vidora"),
		},
		Redis: RedisConfig{
			Addr:     config.GetString("redis.addr", "localhost:6379"),
			Password: config.GetString("redis.password", ""),
			DB:       config.GetInt("redis.db", 0),
		},
	}
	return Cfg, nil
}
```

- [ ] **Step 3: 创建 infra/database.go**

```go
// infra/database.go
package infra

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// NewDB 创建数据库连接
func NewDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db
	return db, nil
}
```

- [ ] **Step 4: 创建 infra/redis.go**

```go
// infra/redis.go
package infra

import (
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// NewRedis 创建 Redis 连接
func NewRedis(cfg *Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	RDB = rdb
	return rdb, nil
}
```

- [ ] **Step 5: 创建 infra/wire.go**

```go
// infra/wire.go
package infra

import "github.com/google/wire"

// ProviderSet infra 层的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewConfig,
	NewDB,
	NewRedis,
)
```

- [ ] **Step 6: 删除旧文件**

```bash
rm infra/init.go
rm infra/dbs.go
```

- [ ] **Step 7: 验证编译**

Run: `go build ./infra/...`
Expected: 无错误

---

## Chunk 2: 迁移 user 模块

### Task 4: 迁移 user 模块 - entity 和 dto

**Files:**
- Create: `module/user/entity.go`
- Create: `module/user/dto.go`
- Create: `module/user/errors.go`
- Delete: `module/user/domain/`

- [ ] **Step 1: 创建 module/user/entity.go**

```go
// modules/user/entity.go
package user

import (
	"time"

	"gorm.io/gorm"
)

// AuditEntity 带审计字段的实体基类
type AuditEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 用户实体
type User struct {
	AuditEntity
	Email    string `gorm:"uniqueIndex;size:255" json:"email"`
	Password string `gorm:"size:255" json:"-"`
	Name     string `gorm:"size:100" json:"name"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
```

- [ ] **Step 2: 创建 module/user/dto.go**

```go
// modules/user/dto.go
package user

// 内部使用的 DTO（handler 层请求/响应）

type CreateReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

type UpdateReq struct {
	Name string `json:"name"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Resp struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
```

- [ ] **Step 3: 创建 module/user/errors.go**

```go
// modules/user/errors.go
package user

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)
```

- [ ] **Step 4: 删除旧的 domain 目录**

```bash
rm -rf modules/user/domain/
```

- [ ] **Step 5: 验证编译**

Run: `go build ./module/user/...`
Expected: 可能提示缺少其他文件，entity/dto/errors 部分无错误

---

### Task 5: 迁移 user 模块 - repository

**Files:**
- Create: `module/user/repository.go`
- Delete: `module/user/repository/`

- [ ] **Step 1: 创建 module/user/repository.go**

```go
// modules/user/repository.go
package user

import (
	"context"

	"gorm.io/gorm"
)

// Repository 用户仓储
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建用户仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建用户
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID 根据 ID 查找用户
func (r *Repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDs 根据 ID 列表查找用户
func (r *Repository) FindByIDs(ctx context.Context, ids []uint) ([]User, error) {
	var users []User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// Update 更新用户
func (r *Repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
```

- [ ] **Step 2: 删除旧的 repository 目录**

```bash
rm -rf modules/user/repository/
```

---

### Task 6: 迁移 user 模块 - service

**Files:**
- Create: `module/user/service.go`
- Delete: `module/user/service/`

- [ ] **Step 1: 创建 module/user/service.go**

```go
// modules/user/service.go
package user

import (
	"context"
	"errors"

	"vidora-api/contract"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 确保实现 contract.UserService 接口
var _ contract.UserService = (*Service)(nil)

// Service 用户服务
type Service struct {
	repo *Repository
}

// NewService 创建用户服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create 创建用户
func (s *Service) Create(ctx context.Context, email, password string) (*contract.UserDTO, error) {
	// 检查邮箱是否已存在
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toDTO(user), nil
}

// Get 获取用户
func (s *Service) Get(ctx context.Context, userID uint) (*contract.UserDTO, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toDTO(user), nil
}

// GetByEmail 根据邮箱获取用户
func (s *Service) GetByEmail(ctx context.Context, email string) (*contract.UserDTO, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toDTO(user), nil
}

// GetByIDs 批量获取用户
func (s *Service) GetByIDs(ctx context.Context, ids []uint) ([]contract.UserDTO, error) {
	users, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.UserDTO, len(users))
	for i, u := range users {
		dtos[i] = *toDTO(&u)
	}
	return dtos, nil
}

// ValidateLogin 验证登录
func (s *Service) ValidateLogin(ctx context.Context, email, password string) (*contract.UserDTO, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return toDTO(user), nil
}

// toDTO 转换为 DTO
func toDTO(user *User) *contract.UserDTO {
	return &contract.UserDTO{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Password: user.Password,
	}
}
```

- [ ] **Step 2: 删除旧的 service 目录**

```bash
rm -rf modules/user/service/
```

---

### Task 7: 迁移 user 模块 - handler

**Files:**
- Create: `module/user/handler.go`
- Delete: `module/user/handler/`

- [ ] **Step 1: 创建 module/user/handler.go**

```go
// modules/user/handler.go
package user

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// Handler 用户 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建用户处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	// User 模块不直接暴露 HTTP 接口
	// 用户相关操作通过 Auth 模块进行
}
```

- [ ] **Step 2: 删除旧的 handler 目录**

```bash
rm -rf modules/user/handler/
```

---

### Task 8: 迁移 user 模块 - module 和 wire

**Files:**
- Create: `module/user/module.go`
- Create: `module/user/wire.go`
- Delete: `module/user/module.go` (旧文件)

- [ ] **Step 1: 创建 module/user/module.go**

```go
// modules/user/modules.go
package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module 用户模块
type Module struct {
	Service *Service
	handler *Handler
}

// NewModule 创建用户模块
func NewModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

// Routes 注册路由
func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
```

- [ ] **Step 2: 创建 module/user/wire.go**

```go
// modules/user/wire.go
package user

import "github.com/google/wire"

// ProviderSet 用户模块的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewRepository,
	NewService,
	NewHandler,
	NewModule,
)
```

- [ ] **Step 3: 验证编译**

Run: `go build ./module/user/...`
Expected: 无错误

---

## Chunk 3: 迁移其他模块

### Task 9: 迁移 category 模块

**Files:**
- Create: `module/category/entity.go`
- Create: `module/category/dto.go`
- Create: `module/category/errors.go`
- Create: `module/category/repository.go`
- Create: `module/category/service.go`
- Create: `module/category/handler.go`
- Create: `module/category/module.go`
- Create: `module/category/wire.go`
- Delete: `module/category/domain/`
- Delete: `module/category/repository/`
- Delete: `module/category/service/`
- Delete: `module/category/handler/`
- Delete: `module/category/module.go` (旧)

- [ ] **Step 1: 读取现有 category 模块代码**

Run: `find module/category -name "*.go" -exec cat {} \;`

- [ ] **Step 2: 创建 module/category/entity.go**

```go
// modules/category/entity.go
package category

import (
	"time"

	"gorm.io/gorm"
)

// AuditEntity 带审计字段的实体基类
type AuditEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category 分类实体
type Category struct {
	AuditEntity
	Name        string `gorm:"size:100" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
```

- [ ] **Step 3: 创建 module/category/dto.go**

```go
// modules/category/dto.go
package category

type CreateReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Resp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
```

- [ ] **Step 4: 创建 module/category/errors.go**

```go
// modules/category/errors.go
package category

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")
)
```

- [ ] **Step 5: 创建 module/category/repository.go**

```go
// modules/category/repository.go
package category

import (
	"context"

	"gorm.io/gorm"
)

// Repository 分类仓储
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建分类仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建分类
func (r *Repository) Create(ctx context.Context, cat *Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

// FindByID 根据 ID 查找分类
func (r *Repository) FindByID(ctx context.Context, id uint) (*Category, error) {
	var cat Category
	err := r.db.WithContext(ctx).First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// FindAll 获取所有分类
func (r *Repository) FindAll(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := r.db.WithContext(ctx).Find(&categories).Error
	return categories, err
}

// Update 更新分类
func (r *Repository) Update(ctx context.Context, cat *Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

// Delete 删除分类
func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Category{}, id).Error
}
```

- [ ] **Step 6: 创建 module/category/service.go**

```go
// modules/category/service.go
package category

import (
	"context"
	"errors"

	"vidora-api/contract"

	"gorm.io/gorm"
)

// 确保实现 contract.CategoryService 接口
var _ contract.CategoryService = (*Service)(nil)

// Service 分类服务
type Service struct {
	repo *Repository
}

// NewService 创建分类服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get 获取分类
func (s *Service) Get(ctx context.Context, id uint) (*contract.CategoryDTO, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return toDTO(cat), nil
}

// Create 创建分类
func (s *Service) Create(ctx context.Context, name, description string) (*Resp, error) {
	cat := &Category{
		Name:        name,
		Description: description,
	}
	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return toResp(cat), nil
}

// Update 更新分类
func (s *Service) Update(ctx context.Context, id uint, name, description string) (*Resp, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		cat.Name = name
	}
	if description != "" {
		cat.Description = description
	}
	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return toResp(cat), nil
}

// Delete 删除分类
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// List 获取分类列表
func (s *Service) List(ctx context.Context) ([]Resp, error) {
	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	resps := make([]Resp, len(categories))
	for i, c := range categories {
		resps[i] = *toResp(&c)
	}
	return resps, nil
}

func toDTO(cat *Category) *contract.CategoryDTO {
	return &contract.CategoryDTO{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
	}
}

func toResp(cat *Category) *Resp {
	return &Resp{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
	}
}
```

- [ ] **Step 7: 创建 module/category/handler.go**

```go
// modules/category/handler.go
package category

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// Handler 分类 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建分类处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/categories", h.Create)
	r.GET("/categories", h.List)
	r.GET("/categories/:id", h.Get)
	r.PUT("/categories/:id", h.Update)
	r.DELETE("/categories/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req.Name, req.Description)
	helper.Respond(c, err, resp)
}

func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resp, err := h.svc.Get(c.Request.Context(), uint(id))
	helper.Respond(c, err, resp)
}

func (h *Handler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context())
	helper.Respond(c, err, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), uint(id), req.Name, req.Description)
	helper.Respond(c, err, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := h.svc.Delete(c.Request.Context(), uint(id))
	helper.Respond(c, err, nil)
}
```

- [ ] **Step 8: 创建 module/category/module.go**

```go
// modules/category/modules.go
package category

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module 分类模块
type Module struct {
	Service *Service
	handler *Handler
}

// NewModule 创建分类模块
func NewModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

// Routes 注册路由
func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
```

- [ ] **Step 9: 创建 module/category/wire.go**

```go
// modules/category/wire.go
package category

import "github.com/google/wire"

// ProviderSet 分类模块的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewRepository,
	NewService,
	NewHandler,
	NewModule,
)
```

- [ ] **Step 10: 删除旧目录**

```bash
rm -rf modules/category/domain/
rm -rf modules/category/repository/
rm -rf modules/category/service/
rm -rf modules/category/handler/
rm -f modules/category/modules.go
```

- [ ] **Step 11: 验证编译**

Run: `go build ./module/category/...`
Expected: 无错误

---

### Task 10: 迁移 tag 模块

**Files:**
- Create: `module/tag/entity.go`
- Create: `module/tag/dto.go`
- Create: `module/tag/errors.go`
- Create: `module/tag/repository.go`
- Create: `module/tag/service.go`
- Create: `module/tag/handler.go`
- Create: `module/tag/module.go`
- Create: `module/tag/wire.go`
- Delete: `module/tag/domain/`, `module/tag/repository/`, `module/tag/service/`, `module/tag/handler/`

- [ ] **Step 1: 读取现有 tag 模块代码**

Run: `find module/tag -name "*.go" -exec cat {} \;`

- [ ] **Step 2: 创建 module/tag/entity.go**

```go
// modules/tag/entity.go
package tag

import (
	"time"

	"gorm.io/gorm"
)

// AuditEntity 带审计字段的实体基类
type AuditEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Tag 标签实体
type Tag struct {
	AuditEntity
	Name string `gorm:"size:50" json:"name"`
}

// VideoTag 视频标签关联
type VideoTag struct {
	VideoID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

func (VideoTag) TableName() string {
	return "video_tags"
}
```

- [ ] **Step 3: 创建 module/tag/dto.go**

```go
// modules/tag/dto.go
package tag

type CreateReq struct {
	Name string `json:"name" binding:"required"`
}

type Resp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
```

- [ ] **Step 4: 创建 module/tag/errors.go**

```go
// modules/tag/errors.go
package tag

import "errors"

var (
	ErrTagNotFound      = errors.New("tag not found")
	ErrInvalidTagIDs    = errors.New("invalid tag IDs")
)
```

- [ ] **Step 5: 创建 module/tag/repository.go**

```go
// modules/tag/repository.go
package tag

import (
	"context"

	"gorm.io/gorm"
)

// Repository 标签仓储
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建标签仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建标签
func (r *Repository) Create(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// FindByID 根据 ID 查找标签
func (r *Repository) FindByID(ctx context.Context, id uint) (*Tag, error) {
	var tag Tag
	err := r.db.WithContext(ctx).First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindByIDs 根据 ID 列表查找标签
func (r *Repository) FindByIDs(ctx context.Context, ids []uint) ([]Tag, error) {
	var tags []Tag
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}

// FindAll 获取所有标签
func (r *Repository) FindAll(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := r.db.WithContext(ctx).Find(&tags).Error
	return tags, err
}

// ExistsByIDs 检查 ID 列表是否都存在
func (r *Repository) ExistsByIDs(ctx context.Context, ids []uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Tag{}).Where("id IN ?", ids).Count(&count).Error
	return count == int64(len(ids)), err
}

// SyncVideoTags 同步视频标签
func (r *Repository) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的关联
		if err := tx.Where("video_id = ?", videoID).Delete(&VideoTag{}).Error; err != nil {
			return err
		}

		// 创建新的关联
		if len(tagIDs) == 0 {
			return nil
		}

		videoTags := make([]VideoTag, len(tagIDs))
		for i, tagID := range tagIDs {
			videoTags[i] = VideoTag{VideoID: videoID, TagID: tagID}
		}
		return tx.Create(&videoTags).Error
	})
}

// Delete 删除标签
func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Tag{}, id).Error
}
```

- [ ] **Step 6: 创建 module/tag/service.go**

```go
// modules/tag/service.go
package tag

import (
	"context"

	"vidora-api/contract"
)

// 确保实现接口
var _ contract.TagService = (*Service)(nil)
var _ contract.TagRepository = (*Repository)(nil)

// Service 标签服务
type Service struct {
	repo *Repository
}

// NewService 创建标签服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ValidateTagIDs 验证标签 ID 列表
func (s *Service) ValidateTagIDs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	exists, err := s.repo.ExistsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidTagIDs
	}
	return nil
}

// Create 创建标签
func (s *Service) Create(ctx context.Context, name string) (*Resp, error) {
	tag := &Tag{Name: name}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}
	return toResp(tag), nil
}

// Get 获取标签
func (s *Service) Get(ctx context.Context, id uint) (*Resp, error) {
	tag, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(tag), nil
}

// List 获取标签列表
func (s *Service) List(ctx context.Context) ([]Resp, error) {
	tags, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	resps := make([]Resp, len(tags))
	for i, t := range tags {
		resps[i] = *toResp(&t)
	}
	return resps, nil
}

// Delete 删除标签
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func toResp(tag *Tag) *Resp {
	return &Resp{
		ID:   tag.ID,
		Name: tag.Name,
	}
}
```

- [ ] **Step 7: 创建 module/tag/handler.go**

```go
// modules/tag/handler.go
package tag

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// Handler 标签 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建标签处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/tags", h.Create)
	r.GET("/tags", h.List)
	r.GET("/tags/:id", h.Get)
	r.DELETE("/tags/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req.Name)
	helper.Respond(c, err, resp)
}

func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resp, err := h.svc.Get(c.Request.Context(), uint(id))
	helper.Respond(c, err, resp)
}

func (h *Handler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context())
	helper.Respond(c, err, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := h.svc.Delete(c.Request.Context(), uint(id))
	helper.Respond(c, err, nil)
}
```

- [ ] **Step 8: 创建 module/tag/module.go**

```go
// modules/tag/modules.go
package tag

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module 标签模块
type Module struct {
	Service *Service
	Repo    *Repository
	handler *Handler
}

// NewModule 创建标签模块
func NewModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Repo:    repo,
		handler: h,
	}
}

// Routes 注册路由
func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
```

- [ ] **Step 9: 创建 module/tag/wire.go**

```go
// modules/tag/wire.go
package tag

import "github.com/google/wire"

// ProviderSet 标签模块的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewRepository,
	NewService,
	NewHandler,
	NewModule,
)
```

- [ ] **Step 10: 删除旧目录**

```bash
rm -rf modules/tag/domain/
rm -rf modules/tag/repository/
rm -rf modules/tag/service/
rm -rf modules/tag/handler/
rm -f modules/tag/modules.go
```

- [ ] **Step 11: 验证编译**

Run: `go build ./module/tag/...`
Expected: 无错误

---

## Chunk 4: 迁移 video 和其他模块

### Task 11: 迁移 video 模块

**Files:**
- Create: `module/video/entity.go`
- Create: `module/video/dto.go`
- Create: `module/video/errors.go`
- Create: `module/video/repository.go`
- Create: `module/video/service.go`
- Create: `module/video/handler.go`
- Create: `module/video/module.go`
- Create: `module/video/wire.go`
- Delete: `module/video/domain/`, `module/video/repository/`, `module/video/service/`, `module/video/handler/`

- [ ] **Step 1: 读取现有 video 模块代码**

Run: `find module/video -name "*.go" -exec cat {} \;`

- [ ] **Step 2: 创建 module/video/entity.go**

```go
// modules/video/entity.go
package video

import (
	"time"

	"vidora-api/module/category"
	"vidora-api/module/tag"

	"gorm.io/gorm"
)

// AuditEntity 带审计字段的实体基类
type AuditEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Video 视频实体
type Video struct {
	AuditEntity
	Title       string          `gorm:"size:255" json:"title"`
	Description string          `gorm:"type:text" json:"description"`
	CoverURL    string          `gorm:"size:512" json:"coverUrl"`
	VideoURL    string          `gorm:"size:512" json:"videoUrl"`
	CategoryID  uint            `json:"categoryId"`
	Category    *category.Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Duration    int             `json:"duration"`
	Status      int8            `json:"status"` // 0: draft, 1: published
	PlayCount   int             `json:"playCount"`
	Tags        []tag.Tag       `gorm:"many2many:video_tags;" json:"tags,omitempty"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}

// IsPublished 是否已发布
func (v *Video) IsPublished() bool {
	return v.Status == 1
}
```

- [ ] **Step 3: 创建 module/video/dto.go**

```go
// modules/video/dto.go
package video

type CreateReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	VideoURL    string `json:"videoUrl" binding:"required"`
	CategoryID  uint   `json:"categoryId" binding:"required"`
	Duration    int    `json:"duration"`
	TagIDs      []uint `json:"tagIds"`
}

type UpdateReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	VideoURL    string `json:"videoUrl"`
	CategoryID  uint   `json:"categoryId"`
	Duration    int    `json:"duration"`
	Status      *int8  `json:"status"`
	TagIDs      []uint `json:"tagIds"`
}

type ListReq struct {
	CategoryID uint   `form:"categoryId"`
	Status     *int8  `form:"status"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
}

type Resp struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	VideoURL    string `json:"videoUrl"`
	CategoryID  uint   `json:"categoryId"`
	Duration    int    `json:"duration"`
	Status      int8   `json:"status"`
	PlayCount   int    `json:"playCount"`
}

type ListResp struct {
	Total int64  `json:"total"`
	List  []Resp `json:"list"`
}
```

- [ ] **Step 4: 创建 module/video/errors.go**

```go
// modules/video/errors.go
package video

import "errors"

var (
	ErrVideoNotFound     = errors.New("video not found")
	ErrVideoNotPublished = errors.New("video not published")
)
```

- [ ] **Step 5: 创建 module/video/repository.go**

```go
// modules/video/repository.go
package video

import (
	"context"

	"gorm.io/gorm"
)

// Repository 视频仓储
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建视频仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建视频
func (r *Repository) Create(ctx context.Context, video *Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

// FindByID 根据 ID 查找视频
func (r *Repository) FindByID(ctx context.Context, id uint) (*Video, error) {
	var video Video
	err := r.db.WithContext(ctx).First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetWithDetails 获取视频详情（包含分类和标签）
func (r *Repository) GetWithDetails(ctx context.Context, id uint) (*Video, error) {
	var video Video
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// Update 更新视频
func (r *Repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Video{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除视频
func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Video{}, id).Error
}

// ExistsByID 检查视频是否存在
func (r *Repository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Video{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ListWithFilter 带过滤条件的列表查询
func (r *Repository) ListWithFilter(ctx context.Context, categoryID uint, status *int8, keyword string, page, pageSize int) ([]Video, int64, error) {
	query := r.db.WithContext(ctx).Model(&Video{})

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var videos []Video
	offset := (page - 1) * pageSize
	err := query.Preload("Category").Preload("Tags").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&videos).Error

	return videos, total, err
}

// IncrementPlayCount 增加播放次数
func (r *Repository) IncrementPlayCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&Video{}).Where("id = ?", id).
		UpdateColumn("play_count", gorm.Expr("play_count + 1")).Error
}
```

- [ ] **Step 6: 创建 module/video/service.go**

```go
// modules/video/service.go
package video

import (
	"context"
	"errors"

	"vidora-api/contract"

	"gorm.io/gorm"
)

// 确保实现接口
var _ contract.VideoService = (*Service)(nil)

// Service 视频服务
type Service struct {
	repo    *Repository
	catSvc  contract.CategoryService
	tagSvc  contract.TagService
	tagRepo contract.TagRepository
}

// NewService 创建视频服务
func NewService(
	repo *Repository,
	catSvc contract.CategoryService,
	tagSvc contract.TagService,
	tagRepo contract.TagRepository,
) *Service {
	return &Service{
		repo:    repo,
		catSvc:  catSvc,
		tagSvc:  tagSvc,
		tagRepo: tagRepo,
	}
}

// Create 创建视频
func (s *Service) Create(ctx context.Context, req contract.CreateVideoReq) (*contract.VideoDTO, error) {
	// 验证分类
	_, err := s.catSvc.Get(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	// 验证标签
	if err := s.tagSvc.ValidateTagIDs(ctx, req.TagIDs); err != nil {
		return nil, err
	}

	video := &Video{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		Status:      0,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return nil, err
	}

	if len(req.TagIDs) > 0 && s.tagRepo != nil {
		_ = s.tagRepo.SyncVideoTags(ctx, video.ID, req.TagIDs)
	}

	return toDTO(video), nil
}

// Update 更新视频
func (s *Service) Update(ctx context.Context, id uint, req contract.UpdateVideoReq) (*contract.VideoDTO, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CoverURL != "" {
		updates["cover_url"] = req.CoverURL
	}
	if req.VideoURL != "" {
		updates["video_url"] = req.VideoURL
	}
	if req.CategoryID > 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Duration > 0 {
		updates["duration"] = req.Duration
	}

	if len(updates) > 0 {
		s.repo.Update(ctx, id, updates)
	}

	if req.TagIDs != nil && s.tagRepo != nil {
		s.tagRepo.SyncVideoTags(ctx, id, req.TagIDs)
	}

	video, err := s.repo.GetWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDTO(video), nil
}

// Delete 删除视频
func (s *Service) Delete(ctx context.Context, id uint) error {
	exists, _ := s.repo.ExistsByID(ctx, id)
	if !exists {
		return ErrVideoNotFound
	}
	return s.repo.Delete(ctx, id)
}

// Get 获取视频
func (s *Service) Get(ctx context.Context, id uint) (*contract.VideoDTO, error) {
	video, err := s.repo.GetWithDetails(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVideoNotFound
	}
	return toDTO(video), err
}

// List 获取视频列表
func (s *Service) List(ctx context.Context, req contract.VideoListReq) (*contract.VideoListDTO, error) {
	videos, total, err := s.repo.ListWithFilter(ctx, req.CategoryID, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.VideoDTO, len(videos))
	for i, v := range videos {
		dtos[i] = *toDTO(&v)
	}

	return &contract.VideoListDTO{Total: total, List: dtos}, nil
}

// GetPublished 获取已发布视频
func (s *Service) GetPublished(ctx context.Context, id uint) (*contract.VideoDTO, error) {
	video, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if video.Status != 1 {
		return nil, ErrVideoNotPublished
	}
	s.repo.IncrementPlayCount(ctx, id)
	return video, nil
}

// ListPublished 获取已发布视频列表
func (s *Service) ListPublished(ctx context.Context, categoryID uint, page, pageSize int) (*contract.VideoListDTO, error) {
	status := int8(1)
	videos, total, err := s.repo.ListWithFilter(ctx, categoryID, &status, "", page, pageSize)
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.VideoDTO, len(videos))
	for i, v := range videos {
		dtos[i] = *toDTO(&v)
	}

	return &contract.VideoListDTO{Total: total, List: dtos}, nil
}

func toDTO(video *Video) *contract.VideoDTO {
	return &contract.VideoDTO{
		ID:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		CoverURL:    video.CoverURL,
		VideoURL:    video.VideoURL,
		CategoryID:  video.CategoryID,
		Duration:    video.Duration,
		Status:      video.Status,
		PlayCount:   video.PlayCount,
	}
}
```

- [ ] **Step 7: 创建 module/video/handler.go**

```go
// modules/video/handler.go
package video

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"vidora-api/contract"
	"vidora-api/server/helper"
)

// Handler 视频 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建视频处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/videos", h.Create)
	r.GET("/videos", h.List)
	r.GET("/videos/:id", h.Get)
	r.PUT("/videos/:id", h.Update)
	r.DELETE("/videos/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	contractReq := contract.CreateVideoReq{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		TagIDs:      req.TagIDs,
	}

	dto, err := h.svc.Create(c.Request.Context(), contractReq)
	helper.Respond(c, err, dto)
}

func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	dto, err := h.svc.Get(c.Request.Context(), uint(id))
	helper.Respond(c, err, dto)
}

func (h *Handler) List(c *gin.Context) {
	var req ListReq
	if !helper.MustBindQuery(c, &req) {
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	contractReq := contract.VideoListReq{
		CategoryID: req.CategoryID,
		Status:     req.Status,
		Keyword:    req.Keyword,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	dto, err := h.svc.List(c.Request.Context(), contractReq)
	helper.Respond(c, err, dto)
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	contractReq := contract.UpdateVideoReq{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		Status:      req.Status,
		TagIDs:      req.TagIDs,
	}

	dto, err := h.svc.Update(c.Request.Context(), uint(id), contractReq)
	helper.Respond(c, err, dto)
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := h.svc.Delete(c.Request.Context(), uint(id))
	helper.Respond(c, err, nil)
}
```

- [ ] **Step 8: 创建 module/video/module.go**

```go
// modules/video/modules.go
package video

import (
	"github.com/gin-gonic/gin"
	"vidora-api/contract"
	"gorm.io/gorm"
)

// Module 视频模块
type Module struct {
	Service *Service
	handler *Handler
}

// NewModule 创建视频模块
func NewModule(
	db *gorm.DB,
	catSvc contract.CategoryService,
	tagSvc contract.TagService,
	tagRepo contract.TagRepository,
) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, catSvc, tagSvc, tagRepo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

// Routes 注册路由
func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
```

- [ ] **Step 9: 创建 module/video/wire.go**

```go
// modules/video/wire.go
package video

import "github.com/google/wire"

// ProviderSet 视频模块的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewRepository,
	NewService,
	NewHandler,
	NewModule,
)
```

- [ ] **Step 10: 删除旧目录**

```bash
rm -rf modules/video/domain/
rm -rf modules/video/repository/
rm -rf modules/video/service/
rm -rf modules/video/handler/
rm -f modules/video/modules.go
```

- [ ] **Step 11: 验证编译**

Run: `go build ./module/video/...`
Expected: 可能需要修复 entity.go 中的导入

---

### Task 12: 迁移 auth 和 transcode 模块

**Files:**
- 迁移 `module/auth/`
- 迁移 `module/transcode/`

- [ ] **Step 1: 迁移 auth 模块**

按照 user 模块的相同模式迁移，创建 entity.go, dto.go, errors.go, repository.go, service.go, handler.go, module.go, wire.go

- [ ] **Step 2: 迁移 transcode 模块**

按照 user 模块的相同模式迁移

- [ ] **Step 3: 验证编译**

Run: `go build ./module/...`
Expected: 无错误

---

## Chunk 5: Wire 集成和最终验证

### Task 13: 创建 module/wire.go

**Files:**
- Create: `module/wire.go`

- [ ] **Step 1: 创建 module/wire.go**

```go
// modules/wire.go
package module

import (
	"github.com/google/wire"
	"vidora-api/module/auth"
	"vidora-api/module/category"
	"vidora-api/module/tag"
	"vidora-api/module/transcode"
	"vidora-api/module/user"
	"vidora-api/module/video"
)

// ProviderSet 合并所有模块的 ProviderSet
var ProviderSet = wire.NewSet(
	user.ProviderSet,
	auth.ProviderSet,
	category.ProviderSet,
	tag.ProviderSet,
	video.ProviderSet,
	transcode.ProviderSet,
)
```

---

### Task 14: 创建 server/wire.go

**Files:**
- Create: `server/wire.go`

- [ ] **Step 1: 创建 server/wire.go**

```go
// server/wire.go
package server

import "github.com/google/wire"

// ProviderSet server 层的 Wire ProviderSet
var ProviderSet = wire.NewSet(
	NewApp,
)
```

---

### Task 15: 重构 server/app.go

**Files:**
- Modify: `server/app.go`

- [ ] **Step 1: 重构 server/app.go**

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

	"vidora-api/infra"
	"vidora-api/module/auth"
	"vidora-api/module/category"
	"vidora-api/module/tag"
	"vidora-api/module/transcode"
	"vidora-api/module/user"
	"vidora-api/module/video"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/ext/logx"
	"github.com/lpphub/goweb/monitor"
)

// App 应用
type App struct {
	engine   *gin.Engine
	user     *user.Module
	auth     *auth.Module
	category *category.Module
	tag      *tag.Module
	video    *video.Module
	transcode *transcode.Module
}

// NewApp 创建应用
func NewApp(
	user *user.Module,
	auth *auth.Module,
	category *category.Module,
	tag *tag.Module,
	video *video.Module,
	transcode *transcode.Module,
) *App {
	return &App{
		engine:    gin.New(),
		user:      user,
		auth:      auth,
		category:  category,
		tag:       tag,
		video:     video,
		transcode: transcode,
	}
}

// Run 启动应用
func (a *App) Run() {
	a.setupRouter()
	a.run()
}

func (a *App) setupRouter() {
	r := a.engine

	r.Use(gin.Recovery())
	r.Use(logx.GinAccessLog(logx.WithSkipPaths("/metrics", "/health")))

	monitor.RegisterMetrics(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册模块路由
	a.user.Routes(r.Group(""))
	a.auth.Routes(r.Group(""))
	a.category.Routes(r.Group(""))
	a.tag.Routes(r.Group(""))
	a.video.Routes(r.Group(""))
	a.transcode.Routes(r.Group(""))
}

func (a *App) run() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", infra.Cfg.Server.Port),
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

---

### Task 16: 创建 cmd/server/wire.go

**Files:**
- Create: `cmd/server/wire.go`

- [ ] **Step 1: 创建 cmd/server/wire.go**

```go
//go:build wireinject

package main

import (
	"github.com/google/wire"
	"vidora-api/infra"
	"vidora-api/module"
	"vidora-api/server"
)

// InitializeApp Wire 生成的初始化函数
func InitializeApp() (*server.App, func(), error) {
	wire.Build(
		infra.ProviderSet,
		module.ProviderSet,
		server.ProviderSet,
	)
	return nil, nil, nil
}
```

---

### Task 17: 更新 cmd/server/main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 更新 cmd/server/main.go**

```go
// cmd/server/main.go
package main

import "log"

func main() {
	app, cleanup, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	app.Run()
}
```

---

### Task 18: 生成 Wire 代码并验证

**Files:**
- Generate: `cmd/server/wire_gen.go`

- [ ] **Step 1: 安装 wire 工具**

Run: `go install github.com/google/wire/cmd/wire@latest`

- [ ] **Step 2: 生成 wire_gen.go**

Run: `wire ./cmd/server/`
Expected: 生成 `cmd/server/wire_gen.go`

- [ ] **Step 3: 验证编译**

Run: `go build ./cmd/server/...`
Expected: 无错误

- [ ] **Step 4: 运行应用测试**

Run: `go run ./cmd/server/`
Expected: 应用正常启动

---

### Task 19: 清理和最终验证

**Files:**
- Delete: 旧文件和目录

- [ ] **Step 1: 删除剩余旧文件**

```bash
# 删除 pkg/core 目录（如果还有）
rm -rf pkg/core/
```

- [ ] **Step 2: 验证完整编译**

Run: `go build ./...`
Expected: 无错误

- [ ] **Step 3: 验证目录结构**

Run: `tree -L 3 -d`
Expected: 目录结构符合设计

- [ ] **Step 4: 提交变更**

```bash
git add .
git commit -m "refactor: 重构项目架构为模块化结构

- 创建 contract 包定义模块间接口
- 重构 infra 包并添加 Wire 支持
- 迁移所有模块为扁平化结构
- 使用 Wire 管理依赖注入
- 删除 biz/ 和 pkg/core/ 目录"
```

---

## 验收标准

- [ ] 目录结构符合设计文档
- [ ] 所有模块编译无错误
- [ ] Wire 代码生成成功
- [ ] 应用可以正常启动
- [ ] 模块间无循环依赖
- [ ] 各模块 HTTP 接口正常响应