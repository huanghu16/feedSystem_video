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
	ErrUserAlreadyExists = errors.New("用户名已存在")
	ErrUserNotFound      = errors.New("用户不存在")
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
		return nil, errors.New("密码错误")
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

// GetProfile 获取用户资料
func (s *Service) GetProfile(userID uint) (*ProfileResponse, error) {
	profile, err := s.repo.GetProfileWithStats(userID)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
