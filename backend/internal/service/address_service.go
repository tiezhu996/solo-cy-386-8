package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
)

// AddressService 收货地址业务服务。
type AddressService struct {
	repo   repository.AddressRepository
	logger *slog.Logger
}

// NewAddressService 构造收货地址服务。
func NewAddressService(repo repository.AddressRepository, logger *slog.Logger) *AddressService {
	return &AddressService{repo: repo, logger: logger}
}

// Create 新增收货地址（默认地址互斥）。
func (s *AddressService) Create(userID uint, req dto.AddressCreateRequest) (*model.Address, error) {
	if req.IsDefault {
		if err := s.repo.ClearDefault(userID); err != nil {
			return nil, fmt.Errorf("clear default address user=%d: %w", userID, err)
		}
	}
	addr := &model.Address{
		UserID:       userID,
		ReceiverName: req.ReceiverName,
		Phone:        req.Phone,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Detail:       req.Detail,
		IsDefault:    req.IsDefault,
	}
	if err := s.repo.Create(addr); err != nil {
		return nil, fmt.Errorf("create address user=%d: %w", userID, err)
	}
	s.logger.Info(constants.LogAddressCreated, "user_id", userID, "address_id", addr.ID, "receiver", addr.ReceiverName)
	return addr, nil
}

// Update 更新收货地址。
func (s *AddressService) Update(userID, addressID uint, req dto.AddressUpdateRequest) (*model.Address, error) {
	addr, err := s.repo.GetByIDAndUser(addressID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeAddressNotFound, "地址更新失败：地址 id="+fmt.Sprint(addressID)+" 不存在或不属于用户 id="+fmt.Sprint(userID), err)
		}
		return nil, fmt.Errorf("get address %d for update: %w", addressID, err)
	}
	if req.ReceiverName != nil {
		addr.ReceiverName = *req.ReceiverName
	}
	if req.Phone != nil {
		addr.Phone = *req.Phone
	}
	if req.Province != nil {
		addr.Province = *req.Province
	}
	if req.City != nil {
		addr.City = *req.City
	}
	if req.District != nil {
		addr.District = *req.District
	}
	if req.Detail != nil {
		addr.Detail = *req.Detail
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := s.repo.ClearDefault(userID); err != nil {
			return nil, fmt.Errorf("clear default address user=%d: %w", userID, err)
		}
		addr.IsDefault = true
	}
	if err := s.repo.Update(addr); err != nil {
		return nil, fmt.Errorf("update address %d: %w", addressID, err)
	}
	s.logger.Info(constants.LogAddressUpdated, "user_id", userID, "address_id", addressID, "receiver", addr.ReceiverName)
	return addr, nil
}

// Delete 删除收货地址。
func (s *AddressService) Delete(userID, addressID uint) error {
	if err := s.repo.Delete(addressID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeAddressNotFound, "地址删除失败：地址 id="+fmt.Sprint(addressID)+" 不存在或不属于用户 id="+fmt.Sprint(userID), err)
		}
		return fmt.Errorf("delete address %d: %w", addressID, err)
	}
	s.logger.Info(constants.LogAddressDeleted, "user_id", userID, "address_id", addressID)
	return nil
}

// List 查询我的收货地址列表。
func (s *AddressService) List(userID uint) ([]model.Address, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses user=%d: %w", userID, err)
	}
	return list, nil
}

// Get 查询单个收货地址。
func (s *AddressService) Get(userID, addressID uint) (*model.Address, error) {
	addr, err := s.repo.GetByIDAndUser(addressID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeAddressNotFound, "地址查询失败：地址 id="+fmt.Sprint(addressID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get address %d: %w", addressID, err)
	}
	return addr, nil
}
