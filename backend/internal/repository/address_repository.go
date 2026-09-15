package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// AddressRepository 收货地址仓储接口。
type AddressRepository interface {
	Create(addr *model.Address) error
	GetByID(id uint) (*model.Address, error)
	GetByIDAndUser(id, userID uint) (*model.Address, error)
	GetByIDAndUserTx(tx *gorm.DB, id, userID uint) (*model.Address, error)
	ListByUser(userID uint) ([]model.Address, error)
	Update(addr *model.Address) error
	Delete(id, userID uint) error
	ClearDefault(userID uint) error
}

type addressRepo struct {
	db *gorm.DB
}

// NewAddressRepository 构造收货地址仓储。
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepo{db: db}
}

func (r *addressRepo) Create(addr *model.Address) error {
	if err := r.db.Create(addr).Error; err != nil {
		return fmt.Errorf("create address: %w", err)
	}
	return nil
}

func (r *addressRepo) GetByID(id uint) (*model.Address, error) {
	var a model.Address
	err := r.db.First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get address by id %d: %w", id, err)
	}
	return &a, nil
}

func (r *addressRepo) GetByIDAndUser(id, userID uint) (*model.Address, error) {
	return r.getByIDAndUser(r.db, id, userID)
}

// GetByIDAndUserTx 在指定事务中查询地址（下单事务内使用，保持同一连接）。
func (r *addressRepo) GetByIDAndUserTx(tx *gorm.DB, id, userID uint) (*model.Address, error) {
	return r.getByIDAndUser(tx, id, userID)
}

func (r *addressRepo) getByIDAndUser(db *gorm.DB, id, userID uint) (*model.Address, error) {
	var a model.Address
	err := db.Where("id = ? AND user_id = ?", id, userID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get address id=%d user=%d: %w", id, userID, err)
	}
	return &a, nil
}

func (r *addressRepo) ListByUser(userID uint) ([]model.Address, error) {
	var list []model.Address
	if err := r.db.Where("user_id = ?", userID).Order("is_default DESC, id DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list addresses user=%d: %w", userID, err)
	}
	return list, nil
}

func (r *addressRepo) Update(addr *model.Address) error {
	if err := r.db.Save(addr).Error; err != nil {
		return fmt.Errorf("update address %d: %w", addr.ID, err)
	}
	return nil
}

func (r *addressRepo) Delete(id, userID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Address{})
	if res.Error != nil {
		return fmt.Errorf("delete address id=%d user=%d: %w", id, userID, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *addressRepo) ClearDefault(userID uint) error {
	res := r.db.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", false)
	if res.Error != nil {
		return fmt.Errorf("clear default address user=%d: %w", userID, res.Error)
	}
	return nil
}
