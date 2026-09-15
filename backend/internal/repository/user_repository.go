package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// ErrNotFound 仓储层哨兵错误。
var ErrNotFound = errors.New("record not found")

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	List(page, pageSize int) ([]model.User, int64, error)
	Update(user *model.User) error
	UpdateCredit(tx *gorm.DB, id uint, delta int) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}
	return &u, nil
}

func (r *userRepo) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username %s: %w", username, err)
	}
	return &u, nil
}

func (r *userRepo) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := r.db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

func (r *userRepo) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user %d: %w", user.ID, err)
	}
	return nil
}

func (r *userRepo) UpdateCredit(tx *gorm.DB, id uint, delta int) error {
	if tx == nil {
		tx = r.db
	}
	res := tx.Model(&model.User{}).Where("id = ?", id).UpdateColumn("credit_score", gorm.Expr("credit_score + ?", delta))
	if res.Error != nil {
		return fmt.Errorf("update credit of user %d: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
