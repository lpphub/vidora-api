// module/user/model.go
package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Status constants
const (
	StatusActive   int8 = 1
	StatusDisabled int8 = 0
)

// Role constants
const (
	RoleUser  int8 = 0
	RoleAdmin int8 = 1
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"column:name" json:"name"`
	Email     string         `gorm:"column:email" json:"email"`
	Password  string         `gorm:"column:password" json:"-"`
	Avatar    string         `gorm:"column:avatar" json:"avatar"`
	Role      int8           `gorm:"column:role;default:0" json:"role"`
	Status    int8           `gorm:"column:status;default:1" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (*User) TableName() string {
	return "users"
}

// IsActive 用户是否激活
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsAdmin 是否管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UpdateProfile 更新个人资料
func (u *User) UpdateProfile(name, avatar string) {
	if name != "" {
		u.Name = name
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	u.UpdatedAt = time.Now()
}

// ValidatePassword 验证密码
func (u *User) ValidatePassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("password mismatch")
		}
		return err
	}
	return nil
}

// SetPassword 设置密码（加密）
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}