package account

import (
	"feedSystem_video/internal/db"

	"golang.org/x/crypto/bcrypt"
)

// =================用户====================
// Repo 封装 Account 的数据库操作
type Repo struct{}

// NewRepo 创建 Repo 实例
func NewRepo() *Repo {
	return &Repo{}
}

// CreateByUsernameAndPassword 创建用户（密码已哈希）
func (r *Repo) CreateByUsernameAndPassword(username, hashedPassword string) (*Account, error) {
	account := Account{
		Username: username,
		Password: hashedPassword,
	}
	// GORM 的 Create 方法：传入指针，自动填充 ID 和 CreatedAt
	if err := db.DB.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// FindByUsername 根据用户名查找用户
func (r *Repo) FindByUsername(username string) (*Account, error) {
	var account Account
	// First 方法：找不到会返回 ErrRecordNotFound
	result := db.DB.Where("username = ?", username).First(&account)
	if result.Error != nil {
		// 记录不存在 → 返回 nil, nil（不是错误）
		// 其他错误（如数据库挂了）→ 返回 nil, err
		if result.RowsAffected == 0 {
			return nil, nil
		}
		return nil, result.Error
	}
	return &account, nil
}

// HashPassword 对密码进行 bcrypt 哈希
// bcrypt 是专门为密码存储设计的哈希算法，自带盐值，防彩虹表攻击
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10，每次哈希约 100ms，故意慢来防暴力破解
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ================登录====================
// FindByID 根据 ID 查找用户
func (r *Repo) FindByID(id uint) (*Account, error) {
	var account Account
	result := db.DB.First(&account, id)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			return nil, nil
		}
		return nil, result.Error
	}
	return &account, nil
}

// UpdateTokens 更新用户的 token 和 refresh_token
func (r *Repo) UpdateTokens(id uint, token, refreshToken string) error {
	return db.DB.Model(&Account{}).Where("id = ?", id).Updates(map[string]interface{}{
		"token":         token,
		"refresh_token": refreshToken,
	}).Error
}

// CheckPassword 验证密码
// bcrypt.CompareHashAndPassword 会自动比较哈希和盐值
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
