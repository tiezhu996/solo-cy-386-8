package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// FavoriteRepository 收藏仓储接口。
type FavoriteRepository interface {
	Create(fav *model.Favorite) error
	Delete(userID, productID uint) error
	Exists(userID, productID uint) (bool, error)
	ListByUser(userID uint, page, pageSize int) ([]model.Favorite, int64, error)
}

type favoriteRepo struct {
	db *gorm.DB
}

// NewFavoriteRepository 构造收藏仓储。
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepo{db: db}
}

func (r *favoriteRepo) Create(fav *model.Favorite) error {
	if err := r.db.Create(fav).Error; err != nil {
		return fmt.Errorf("create favorite: %w", err)
	}
	return nil
}

func (r *favoriteRepo) Delete(userID, productID uint) error {
	res := r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&model.Favorite{})
	if res.Error != nil {
		return fmt.Errorf("delete favorite user=%d product=%d: %w", userID, productID, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *favoriteRepo) Exists(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check favorite user=%d product=%d: %w", userID, productID, err)
	}
	return count > 0, nil
}

func (r *favoriteRepo) ListByUser(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	var favs []model.Favorite
	var total int64
	q := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}
	if err := q.Preload("Product").Preload("Product.Seller").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&favs).Error; err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	return favs, total, nil
}

// EnsureFavoriteError 兼容 errors.Is 判断。
func EnsureFavoriteError(err error) bool {
	return errors.Is(err, ErrNotFound)
}
