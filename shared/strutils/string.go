package strutils

import (
	"strings"
)

// ExtractNameFromEmail 从邮箱提取名称
func ExtractNameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return email
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
