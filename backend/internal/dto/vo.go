package dto

import (
	"strings"
	"time"

	"github.com/marketpal/marketpal/internal/model"
)

// vo.go 集中 model → 视图对象 转换（handler/service 复用，避免重复代码）。

// FromUser 用户实体转视图对象。
func FromUser(u *model.User) *UserVO {
	return &UserVO{
		ID:          u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Email:       u.Email,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		Role:        u.Role,
		CreditScore: u.CreditScore,
		Status:      u.Status,
		CreatedAt:   formatTime(u.CreatedAt),
	}
}

// FromProduct 商品实体转视图对象。
func FromProduct(p *model.Product, isFavorite bool) ProductVO {
	var seller *UserVO
	if p.Seller != nil {
		seller = FromUser(p.Seller)
	}
	images := []string{}
	if p.Images != "" {
		images = strings.Split(p.Images, ",")
	}
	return ProductVO{
		ID:            p.ID,
		SellerID:      p.SellerID,
		Title:         p.Title,
		Description:   p.Description,
		OriginalPrice: p.OriginalPrice,
		Price:         p.Price,
		Condition:     p.Condition,
		Category:      p.Category,
		Images:        images,
		Status:        p.Status,
		ViewCount:     p.ViewCount,
		FavoriteCount: p.FavoriteCount,
		CreatedAt:     formatTime(p.CreatedAt),
		Seller:        seller,
		IsFavorite:    isFavorite,
	}
}

// FromAddress 地址实体转视图对象。
func FromAddress(a *model.Address) AddressVO {
	return AddressVO{
		ID:           a.ID,
		ReceiverName: a.ReceiverName,
		Phone:        a.Phone,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		Detail:       a.Detail,
		IsDefault:    a.IsDefault,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
