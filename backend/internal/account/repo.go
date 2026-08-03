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

// FindByRefreshToken 根据 refresh_token 查找用户
// 安全要点：空字符串必须直接返回 nil，否则会匹配到从未登录的用户（refresh_token 字段默认 ''）
func (r *Repo) FindByRefreshToken(refreshToken string) (*Account, error) {
	if refreshToken == "" {
		return nil, nil
	}
	var account Account
	result := db.DB.Where("refresh_token = ?", refreshToken).First(&account)
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

// RotateTokens 原子轮换 token（乐观锁 / CAS 思想）
// 仅当库中 refresh_token 仍等于 oldRefreshToken 时才更新成功：
//   - 正常刷新：旧值匹配，RowsAffected=1，旧 refresh_token 立即作废
//   - 并发刷新或 token 重放：旧值已被前一个请求轮换掉，RowsAffected=0，拒绝
//
// 返回 false 表示轮换失败（应视为 refresh_token 无效）
func (r *Repo) RotateTokens(id uint, oldRefreshToken, newToken, newRefreshToken string) (bool, error) {
	result := db.DB.Model(&Account{}).
		Where("id = ? AND refresh_token = ?", id, oldRefreshToken).
		Updates(map[string]interface{}{
			"token":         newToken,
			"refresh_token": newRefreshToken,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// CheckPassword 验证密码
// bcrypt.CompareHashAndPassword 会自动比较哈希和盐值
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetProfileWithStats 获取用户资料和统计信息
func (r *Repo) GetProfileWithStats(userID uint) (*ProfileResponse, error) {
	var account Account // 用户表
	// 查询用户表
	if err := db.DB.First(&account, userID).Error; err != nil {
		return nil, err
	}

	var fansCount int64 // 粉丝数
	db.DB.Table("socials").Where("vlogger_id = ?", userID).Count(&fansCount)

	var followingCount int64 // 关注数
	db.DB.Table("socials").Where("follower_id = ?", userID).Count(&followingCount)

	var videoCount int64 // 视频数
	db.DB.Table("videos").Where("author_id = ?", userID).Count(&videoCount)

	var likesCount int64 // 点赞数
	db.DB.Table("likes").
		Joins("JOIN videos ON likes.video_id = videos.id").
		Where("videos.author_id = ?", userID).
		Distinct("likes.id").
		Count(&likesCount)

	return &ProfileResponse{
		ID:             account.ID,
		Username:       account.Username,
		AvatarURL:      account.AvatarURL,
		Bio:            account.Bio,
		FansCount:      fansCount,
		FollowingCount: followingCount,
		VideoCount:     videoCount,
		LikesCount:     likesCount,
	}, nil
}

// UpdateAvatar 更新用户头像
func (r *Repo) UpdateAvatar(userID uint, avatarURL string) error {
	return db.DB.Model(&Account{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error
}

// UpdatePassword 更新用户密码
func (r *Repo) UpdatePassword(userID uint, hashedPassword string) error {
	return db.DB.Model(&Account{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// UpdateBio 更新用户简介
func (r *Repo) UpdateBio(userID uint, bio string) error {
	return db.DB.Model(&Account{}).Where("id = ?", userID).Update("bio", bio).Error
}
