package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
	"gorm.io/gorm"
)

// UserService 用户业务服务。
type UserService struct {
	repo   repository.UserRepository
	logger *slog.Logger
	secret string
	ttl    time.Duration
}

// NewUserService 构造用户服务（构造器注入）。
func NewUserService(repo repository.UserRepository, logger *slog.Logger, secret string, ttl time.Duration) *UserService {
	return &UserService{repo: repo, logger: logger, secret: secret, ttl: ttl}
}

// Register 注册：校验用户名唯一 → 密码哈希 → 落库。
func (s *UserService) Register(req dto.RegisterRequest) (*model.User, error) {
	if _, err := s.repo.GetByUsername(req.Username); err == nil {
		return nil, util.NewAppError(constants.CodeUserExists, "用户注册失败：用户名 "+req.Username+" 已被占用", nil)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("register user %s: check username: %w", req.Username, err)
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("register user %s: hash password: %w", req.Username, err)
	}
	user := &model.User{
		Username:    req.Username,
		PasswordHash: hash,
		Nickname:    req.Nickname,
		Email:       req.Email,
		Phone:       req.Phone,
		Role:        constants.UserRoleUser,
		CreditScore: 100,
		Status:      "active",
	}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("register user %s: %w", req.Username, err)
	}
	s.logger.Info(constants.LogUserRegistered, "user_id", user.ID, "username", user.Username, "role", user.Role)
	return user, nil
}

// Login 登录：校验凭据并签发 JWT。
func (s *UserService) Login(req dto.LoginRequest) (*model.User, string, error) {
	user, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", util.NewAppError(constants.CodeInvalidCredentials, "用户登录失败：用户名或密码错误", nil)
		}
		return nil, "", fmt.Errorf("login user %s: %w", req.Username, err)
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		return nil, "", util.NewAppError(constants.CodeInvalidCredentials, "用户登录失败：用户名或密码错误", nil)
	}
	if user.Status != "active" {
		return nil, "", util.NewAppError(constants.CodeForbidden, "用户登录失败：账号已被禁用，角色 "+user.Role, nil)
	}
	token, err := util.GenerateToken(s.secret, user.ID, user.Username, user.Role, s.ttl)
	if err != nil {
		return nil, "", fmt.Errorf("login user %s: generate token: %w", req.Username, err)
	}
	s.logger.Info(constants.LogUserLogin, "user_id", user.ID, "username", user.Username, "role", user.Role)
	return user, token, nil
}

// GetProfile 获取个人资料。
func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeUserNotFound, "查询用户资料失败：用户 id="+fmt.Sprint(userID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get profile user %d: %w", userID, err)
	}
	return user, nil
}

// UpdateProfile 更新个人资料。
func (s *UserService) UpdateProfile(userID uint, req dto.UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeUserNotFound, "更新用户资料失败：用户 id="+fmt.Sprint(userID)+" 不存在", err)
		}
		return nil, fmt.Errorf("update profile user %d: %w", userID, err)
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("update profile user %d: %w", userID, err)
	}
	s.logger.Info(constants.LogUserProfileUpdated, "user_id", user.ID, "nickname", user.Nickname)
	return user, nil
}

// ListUsers 管理员分页查询用户列表。
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	p := util.NormalizePage(page, pageSize)
	users, total, err := s.repo.List(p.Page, p.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

// UpdateRole 管理员修改用户角色。
func (s *UserService) UpdateRole(operatorID, targetID uint, role string) (*model.User, error) {
	if !constants.ValidUserRole(role) {
		return nil, util.NewAppError(constants.CodeBadRequest, "修改用户角色失败：角色 "+role+" 非法", nil)
	}
	target, err := s.repo.GetByID(targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeUserNotFound, "修改用户角色失败：目标用户 id="+fmt.Sprint(targetID)+" 不存在", err)
		}
		return nil, fmt.Errorf("update role target %d: %w", targetID, err)
	}
	target.Role = role
	if err := s.repo.Update(target); err != nil {
		return nil, fmt.Errorf("update role target %d: %w", targetID, err)
	}
	s.logger.Info(constants.LogUserRoleUpdated, "operator", operatorID, "target_user", targetID, "role", role)
	return target, nil
}

// UpdateCredit 评价系统调用的信用积分更新（供 ReviewService 复用同一仓储方法）。
func (s *UserService) UpdateCredit(tx *gorm.DB, userID uint, delta int) error {
	return s.repo.UpdateCredit(tx, userID, delta)
}
