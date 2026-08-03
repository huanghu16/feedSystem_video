package account

import (
	"errors"
	"feedSystem_video/internal/auth"
	"feedSystem_video/internal/config"
)

type Service struct {
	repo *Repo
}

// NewService 创建 Service 实例
// 参数 repo 是从外部传入的，这叫"依赖注入"
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// 定义业务错误常量
// 用 errors.New 创建，上层可以 errors.Is(err, ErrUserAlreadyExists) 精确匹配
var (
	ErrUserAlreadyExists   = errors.New("用户名已存在")
	ErrUserNotFound        = errors.New("用户不存在")
	ErrPasswordWrong       = errors.New("密码错误")
	ErrOldPasswordWrong    = errors.New("原密码错误")
	ErrAvatarURLEmpty      = errors.New("头像URL不能为空")
	ErrBioTooLong          = errors.New("简介不能超过256个字符")
	ErrRefreshTokenInvalid = errors.New("refresh token 无效或已过期")
)

// Register 注册新用户
// 流程：查重 → 哈希密码 → 写库 → 返回结果
func (s *Service) Register(req *RegisterRequest) (*RegisterResponse, error) {
	//查用户名是否已存在
	existing, _ := s.repo.FindByUsername(req.Username)
	if existing != nil {
		// 用户名已存在, 返回错误
		return nil, ErrUserAlreadyExists
	}

	//对密码做 bcrypt 哈希
	hashed, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	//写入数据库
	account, err := s.repo.CreateByUsernameAndPassword(req.Username, hashed)
	if err != nil {
		return nil, err
	}

	//返回结果(只返回必要字段)
	return &RegisterResponse{
		ID:       account.ID,
		Username: account.Username,
	}, nil
}

// Login 用户登录
func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	// 第一步：查用户
	account, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrUserNotFound
	}

	// 第二步：验证密码
	if !CheckPassword(account.Password, req.Password) {
		return nil, ErrPasswordWrong
	}

	// 第三步：生成双 Token
	accessToken, refreshToken, err := auth.GenerateTokenPair(account.ID, account.Username)
	if err != nil {
		return nil, err
	}

	// 第四步：保存 token 到数据库（后续会加 Redis 缓存）
	if err := s.repo.UpdateTokens(account.ID, accessToken, refreshToken); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.C.JWT.AccessTTL,
	}, nil
}

// RefreshToken 用 refresh_token 换取新的双 Token
// 安全设计（Refresh Token Rotation）：
//  1. refresh_token 一次性使用：每次刷新都签发新的 refresh_token，旧的立即作废，防窃取重放
//  2. CAS 原子轮换落库：UPDATE ... WHERE refresh_token = old，并发刷新/重放只有一个能成功
func (s *Service) RefreshToken(refreshToken string) (*LoginResponse, error) {
	if refreshToken == "" {
		return nil, ErrRefreshTokenInvalid
	}

	// 第一步：根据 refresh_token 查用户（查不到统一视为无效，不区分原因，防探测）
	account, err := s.repo.FindByRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrRefreshTokenInvalid
	}

	// 第二步：生成新的双 Token
	accessToken, newRefreshToken, err := auth.GenerateTokenPair(account.ID, account.Username)
	if err != nil {
		return nil, err
	}

	// 第三步：CAS 落库——仅当库中仍是旧 refresh_token 时更新成功，保证旧 token 只能用一次
	ok, err := s.repo.RotateTokens(account.ID, refreshToken, accessToken, newRefreshToken)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRefreshTokenInvalid
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    config.C.JWT.AccessTTL,
	}, nil
}

// GetProfile 获取用户资料
func (s *Service) GetProfile(userID uint) (*ProfileResponse, error) {
	profile, err := s.repo.GetProfileWithStats(userID)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// UpdateAvatar 更新用户头像
func (s *Service) UpdateAvatar(userID uint, avatarURL string) error {
	if avatarURL == "" {
		return ErrAvatarURLEmpty
	}
	return s.repo.UpdateAvatar(userID, avatarURL)
}

// ChangePassword 修改用户密码
func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 验证旧密码
	account, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrUserNotFound
	}

	if !CheckPassword(account.Password, oldPassword) {
		return ErrOldPasswordWrong
	}

	// 哈希新密码
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return s.repo.UpdatePassword(userID, hashed)
}

// UpdateBio 更新用户简介
func (s *Service) UpdateBio(userID uint, bio string) error {
	if len(bio) > 256 {
		return ErrBioTooLong
	}
	return s.repo.UpdateBio(userID, bio)
}
