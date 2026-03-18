# Tag Module Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor the tag module to support tag groups with the new API structure per `/home/lsk/projects/vidora/vidora-admin/docs/tag.md`.

**Architecture:** Introduce `TagGroup` model with a one-to-many relationship to `Tag`. The default tag group (ID=0) is virtual and handled specially in code. Tags are organized under groups, and when a group is deleted, its tags move to the default group.

**Tech Stack:** Go 1.25, Gin, GORM, MySQL

---

## Database Schema

### tag_groups table
```sql
CREATE TABLE tag_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at DATETIME NOT NULL
);
```

### Modified tags table
```sql
ALTER TABLE tags ADD COLUMN group_id BIGINT UNSIGNED DEFAULT 0;
ALTER TABLE tags DROP INDEX name;
ALTER TABLE tags ADD UNIQUE INDEX idx_group_name (group_id, name);
```

---

## Task 1: Create TagGroup Model

**Files:**
- Create: `modules/tag/model_group.go`

**Step 1: Write TagGroup model**

```go
package tag

import (
	"time"

	"gorm.io/gorm"
)

type TagGroup struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	Tags      []Tag          `gorm:"foreignKey:GroupID" json:"tagList,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (*TagGroup) TableName() string {
	return "tag_groups"
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/tag/model_group.go
git commit -m "feat(tag): add TagGroup model"
```

---

## Task 2: Update Tag Model with GroupID

**Files:**
- Modify: `modules/tag/model.go:20-35`

**Step 1: Update Tag struct**

Replace the `Tag` struct (lines 20-30) with:

```go
type Tag struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	GroupID    uint           `gorm:"not null;default:0;index:idx_group_name" json:"groupId"`
	Name       string         `gorm:"size:50;not null;index:idx_group_name" json:"name"`
	Type       TagType        `gorm:"default:0;index" json:"type,omitempty"`
	Color      string         `gorm:"size:7" json:"color,omitempty"`
	SortOrder  int            `gorm:"default:0" json:"sortOrder,omitempty"`
	Status     TagStatus      `gorm:"size:10;default:'enabled'" json:"status,omitempty"`
	UsageCount int            `gorm:"-" json:"usageCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
```

**Step 2: Remove IsCategory method**

Delete lines 37-40 (the `IsCategory` method).

**Step 3: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add modules/tag/model.go
git commit -m "feat(tag): add GroupID to Tag model"
```

---

## Task 3: Create TagGroup Repository

**Files:**
- Create: `modules/tag/repository_group.go`

**Step 1: Write TagGroup repository**

```go
package tag

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

type GroupRepository struct {
	*dbx.BaseRepo[TagGroup]
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{
		BaseRepo: dbx.NewBaseRepo[TagGroup](db),
	}
}

func (r *GroupRepository) ListWithTags(ctx context.Context) ([]TagGroup, error) {
	var groups []TagGroup
	err := r.DB().WithContext(ctx).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&groups).Error
	return groups, err
}

func (r *GroupRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&TagGroup{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *GroupRepository) UpdateSortOrders(ctx context.Context, ids []uint) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&TagGroup{}).Where("id = ?", id).Update("sort_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GroupRepository) GetMaxSortOrder(ctx context.Context) (int, error) {
	var maxOrder int
	err := r.DB().WithContext(ctx).Model(&TagGroup{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/tag/repository_group.go
git commit -m "feat(tag): add TagGroup repository"
```

---

## Task 4: Update Tag Repository for Groups

**Files:**
- Modify: `modules/tag/repository.go`

**Step 1: Update List method to filter by group**

Replace `List` method (lines 20-24) with:

```go
func (r *Repository) List(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Order("created_at ASC").Find(&tags).Error
	return tags, err
}

func (r *Repository) ListByGroup(ctx context.Context, groupID uint) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Where("group_id = ?", groupID).Order("created_at ASC").Find(&tags).Error
	return tags, err
}
```

**Step 2: Update ExistsByName to include groupID**

Replace `ExistsByName` method (lines 32-36) with:

```go
func (r *Repository) ExistsByName(ctx context.Context, name string, groupID uint) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&Tag{}).Where("name = ? AND group_id = ?", name, groupID).Count(&count).Error
	return count > 0, err
}
```

**Step 3: Add MoveToGroup method**

Add after `GetUsageCount` method:

```go
func (r *Repository) MoveToGroup(ctx context.Context, fromGroupID, toGroupID uint) error {
	return r.DB().WithContext(ctx).Model(&Tag{}).Where("group_id = ?", fromGroupID).Update("group_id", toGroupID).Error
}
```

**Step 4: Run typecheck**

Run: `go build ./...`
Expected: Errors about `ExistsByName` calls in service.go (will fix next)

**Step 5: Commit**

```bash
git add modules/tag/repository.go
git commit -m "feat(tag): update Tag repository for group support"
```

---

## Task 5: Create TagGroup Service

**Files:**
- Create: `modules/tag/service_group.go`

**Step 1: Write TagGroup service**

```go
package tag

import (
	"context"
	"errors"

	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

type GroupService struct {
	repo     *GroupRepository
	tagRepo  *Repository
}

func NewGroupService(repo *GroupRepository, tagRepo *Repository) *GroupService {
	return &GroupService{repo: repo, tagRepo: tagRepo}
}

type CreateGroupReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type UpdateGroupReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type ReorderGroupsReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

const DefaultGroupID uint = 0

func (s *GroupService) List(ctx context.Context) ([]TagGroup, error) {
	groups, err := s.repo.ListWithTags(ctx)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		for j := range groups[i].Tags {
			count, _ := s.tagRepo.GetUsageCount(ctx, groups[i].Tags[j].ID)
			groups[i].Tags[j].UsageCount = int(count)
		}
	}
	return groups, nil
}

func (s *GroupService) Create(ctx context.Context, req CreateGroupReq) (*TagGroup, error) {
	exists, _ := s.repo.ExistsByName(ctx, req.Name)
	if exists {
		return nil, errs.ErrTagGroupExists
	}
	maxOrder, _ := s.repo.GetMaxSortOrder(ctx)
	group := &TagGroup{
		Name:      req.Name,
		SortOrder: maxOrder + 1,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		return nil, err
	}
	group.Tags = []Tag{}
	return group, nil
}

func (s *GroupService) Update(ctx context.Context, id uint, req UpdateGroupReq) (*TagGroup, error) {
	group, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	if req.Name != group.Name {
		exists, _ := s.repo.ExistsByName(ctx, req.Name)
		if exists {
			return nil, errs.ErrTagGroupExists
		}
		if err := s.repo.Update(ctx, id, map[string]any{"name": req.Name}); err != nil {
			return nil, err
		}
		group.Name = req.Name
	}
	group.Tags, _ = s.tagRepo.ListByGroup(ctx, id)
	return group, nil
}

func (s *GroupService) Delete(ctx context.Context, id uint) error {
	if id == DefaultGroupID {
		return errs.ErrCannotDeleteDefaultGroup
	}
	_, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrTagGroupNotFound
	}
	if err != nil {
		return err
	}
	if err := s.tagRepo.MoveToGroup(ctx, id, DefaultGroupID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) Reorder(ctx context.Context, ids []uint) error {
	return s.repo.UpdateSortOrders(ctx, ids)
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: Errors about undefined errs (will add in next task)

**Step 3: Commit**

```bash
git add modules/tag/service_group.go
git commit -m "feat(tag): add TagGroup service"
```

---

## Task 6: Add New Error Types

**Files:**
- Modify: `shared/errs/errors.go:30-33`

**Step 1: Add tag group errors**

Add after `ErrUnsupportedType` (line 41):

```go
	ErrTagGroupNotFound        = base.NewError(2209, "标签组不存在")
	ErrTagGroupExists          = base.NewError(2210, "标签组已存在")
	ErrCannotDeleteDefaultGroup = base.NewError(2211, "不能删除默认标签组")
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: Errors in tag/service.go (will fix next)

**Step 3: Commit**

```bash
git add shared/errs/errors.go
git commit -m "feat(errs): add tag group error types"
```

---

## Task 7: Update Tag Service for Groups

**Files:**
- Modify: `modules/tag/service.go`

**Step 1: Update CreateTagReq**

Replace `CreateTagReq` struct (lines 23-29) with:

```go
type CreateTagReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type CreateTagInGroupReq struct {
	Name    string `json:"name" binding:"required,max=50"`
	GroupID uint   `json:"-"`
}
```

**Step 2: Update UpdateTagReq**

Replace `UpdateTagReq` struct (lines 31-36) with:

```go
type UpdateTagReq struct {
	Name string `json:"name" binding:"required,max=50"`
}
```

**Step 3: Update Create method**

Replace `Create` method (lines 38-66) with:

```go
func (s *Service) Create(ctx context.Context, req CreateTagInGroupReq) (*Tag, error) {
	exists, _ := s.repo.ExistsByName(ctx, req.Name, req.GroupID)
	if exists {
		return nil, errs.ErrTagExists
	}

	tag := &Tag{
		GroupID: req.GroupID,
		Name:    req.Name,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}
```

**Step 4: Update GetByID method**

Replace `GetByID` method (lines 68-82) with:

```go
func (s *Service) GetByID(ctx context.Context, id uint) (*Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	if err != nil {
		return nil, err
	}
	count, err := s.repo.GetUsageCount(ctx, id)
	if err != nil {
		return nil, err
	}
	tag.UsageCount = int(count)
	return tag, nil
}
```

**Step 5: Remove List method**

Delete `List` method (lines 84-100).

**Step 6: Update Update method**

Replace `Update` method (lines 102-147) with:

```go
func (s *Service) Update(ctx context.Context, id uint, req UpdateTagReq) (*Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	if err != nil {
		return nil, err
	}

	if req.Name != "" && req.Name != tag.Name {
		exists, _ := s.repo.ExistsByName(ctx, req.Name, tag.GroupID)
		if exists {
			return nil, errs.ErrTagExists
		}
		if err := s.repo.Update(ctx, id, map[string]any{"name": req.Name}); err != nil {
			return nil, err
		}
		tag.Name = req.Name
	}

	count, err := s.repo.GetUsageCount(ctx, id)
	if err != nil {
		return nil, err
	}
	tag.UsageCount = int(count)
	return tag, nil
}
```

**Step 7: Keep remaining methods unchanged**

Verify `Delete`, `ExistByIDs`, `GetVideoTags`, `SyncVideoTags` methods remain (lines 149-180).

**Step 8: Run typecheck**

Run: `go build ./...`
Expected: Errors in handler.go (will fix next)

**Step 9: Commit**

```bash
git add modules/tag/service.go
git commit -m "refactor(tag): update Tag service for group support"
```

---

## Task 8: Create TagGroup Handler

**Files:**
- Create: `modules/tag/handler_group.go`

**Step 1: Write TagGroup handler**

```go
package tag

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

type GroupHandler struct {
	svc *GroupService
}

func NewGroupHandler(svc *GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.svc.List(c.Request.Context())
	helper.Respond(c, err, groups)
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req CreateGroupReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	group, err := h.svc.Create(c.Request.Context(), req)
	helper.Respond(c, err, group)
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	var req UpdateGroupReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	group, err := h.svc.Update(c.Request.Context(), id, req)
	helper.Respond(c, err, group)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), id)
	helper.Respond(c, err, gin.H{"success": true})
}

func (h *GroupHandler) Reorder(c *gin.Context) {
	var req ReorderGroupsReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	err := h.svc.Reorder(c.Request.Context(), req.IDs)
	helper.Respond(c, err, gin.H{"success": true})
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/tag/handler_group.go
git commit -m "feat(tag): add TagGroup handler"
```

---

## Task 9: Update Tag Handler

**Files:**
- Modify: `modules/tag/handler.go`

**Step 1: Replace entire handler.go file**

Replace the entire `handler.go` file with:

```go
package tag

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

type Handler struct {
	svc      *Service
	groupSvc *GroupService
}

func NewHandler(svc *Service, groupSvc *GroupService) *Handler {
	return &Handler{svc: svc, groupSvc: groupSvc}
}

func (h *Handler) register(r *gin.RouterGroup) {
	gh := NewGroupHandler(h.groupSvc)
	r.GET("/tag-groups", gh.List)
	r.POST("/tag-groups", gh.Create)
	r.PUT("/tag-groups/:id", gh.Update)
	r.DELETE("/tag-groups/:id", gh.Delete)
	r.PUT("/tag-groups/reorder", gh.Reorder)
	r.POST("/tag-groups/:groupId/tags", h.CreateTag)
	r.PUT("/tag-groups/:groupId/tags/:tagId", h.UpdateTag)
	r.DELETE("/tag-groups/:groupId/tags/:tagId", h.DeleteTag)
}

func (h *Handler) CreateTag(c *gin.Context) {
	groupID, ok := helper.MustParseUintParam(c, "groupId")
	if !ok {
		return
	}
	var req CreateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Create(c.Request.Context(), CreateTagInGroupReq{
		Name:    req.Name,
		GroupID: groupID,
	})
	helper.Respond(c, err, tag)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	tagID, ok := helper.MustParseUintParam(c, "tagId")
	if !ok {
		return
	}
	var req UpdateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Update(c.Request.Context(), tagID, req)
	helper.Respond(c, err, tag)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	tagID, ok := helper.MustParseUintParam(c, "tagId")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), tagID)
	helper.Respond(c, err, gin.H{"success": true})
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/tag/handler.go
git commit -m "refactor(tag): update handler for new API structure"
```

---

## Task 10: Update Module Init

**Files:**
- Modify: `modules/tag/init.go`

**Step 1: Replace entire init.go file**

Replace the entire `init.go` file with:

```go
package tag

import (
	"vidora-api/modules/core"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	Service *Service
	handler *Handler
}

func New(db *gorm.DB) *Module {
	repo := NewRepository(db)
	groupRepo := NewGroupRepository(db)
	svc := NewService(repo)
	groupSvc := NewGroupService(groupRepo, repo)
	h := NewHandler(svc, groupSvc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.register(r)
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/tag/init.go
git commit -m "refactor(tag): update module init for TagGroup support"
```

---

## Task 11: Update Contract Interface

**Files:**
- Modify: `modules/core/contract/tag.go`

**Step 1: Update TagDTO**

Replace `TagDTO` struct (lines 13-17) with:

```go
type TagDTO struct {
	ID      uint
	Name    string
	GroupID uint
}
```

**Step 2: Run typecheck**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add modules/core/contract/tag.go
git commit -m "refactor(contract): update TagDTO with GroupID"
```

---

## Task 12: Build and Verify

**Step 1: Run full build**

Run: `go build ./...`
Expected: No errors

**Step 2: Run tests**

Run: `go test ./...`
Expected: All tests pass (or no tests exist)

**Step 3: Run linter (if configured)**

Run: `go vet ./...`
Expected: No issues

**Step 4: Final commit**

```bash
git add -A
git commit -m "refactor(tag): complete tag module refactoring with tag groups"
```

---

## Summary of API Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | /api/tag-groups | GroupHandler.List | Get all tag groups with tags |
| POST | /api/tag-groups | GroupHandler.Create | Create new tag group |
| PUT | /api/tag-groups/:id | GroupHandler.Update | Update tag group name |
| DELETE | /api/tag-groups/:id | GroupHandler.Delete | Delete tag group (moves tags to default) |
| PUT | /api/tag-groups/reorder | GroupHandler.Reorder | Reorder tag groups |
| POST | /api/tag-groups/:groupId/tags | Handler.CreateTag | Create tag in group |
| PUT | /api/tag-groups/:groupId/tags/:tagId | Handler.UpdateTag | Update tag |
| DELETE | /api/tag-groups/:groupId/tags/:tagId | Handler.DeleteTag | Delete tag |

---

## Database Migration Notes

After deployment, run:

```sql
-- Create tag_groups table
CREATE TABLE IF NOT EXISTS tag_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at DATETIME NOT NULL,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_deleted_at (deleted_at)
);

-- Add group_id to tags
ALTER TABLE tags ADD COLUMN group_id BIGINT UNSIGNED DEFAULT 0;

-- Drop old unique index on name (if exists)
ALTER TABLE tags DROP INDEX name;

-- Add new composite unique index
ALTER TABLE tags ADD UNIQUE INDEX idx_group_name (group_id, name);

-- Add index on group_id
ALTER TABLE tags ADD INDEX idx_group_id (group_id);
```