// modules/user/repository.go
package user

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

// Repository 用户仓储
type Repository struct {
	*dbx.BaseRepo[User]
}

// NewRepository 创建用户仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[User](db),
	}
}

// GetByEmail 根据邮箱获取用户
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.DB().WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmails 根据邮箱列表获取用户
func (r *Repository) GetByEmails(ctx context.Context, emails []string) ([]*User, error) {
	var users []*User
	if err := r.DB().WithContext(ctx).Where("email IN ?", emails).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.DB().WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
