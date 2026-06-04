package service

import (
	"context"
	"errors"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/luhn"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userId int64, number string) (orderUserId int64, err error)
	GetOrdersByUserID(ctx context.Context, userId int64) ([]*model.Order, error)
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userId int64, number string) error {

	isValid := luhn.Validate(number)
	if isValid != 0 {
		if isValid < 0 {
			return errs.ErrOrderNumberInvalid
		}
		return errs.ErrOrderNumberUnprocessable
	}

	orderUserId, err := s.repo.CreateOrder(ctx, userId, number)
	if err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyExists) {
			if orderUserId != userId {
				return errs.ErrOrderConflict
			}
			return errs.ErrOrderAlreadyUploaded
		}

		return err
	}

	return nil
}

func (s *OrderService) GetOrdersByUserID(ctx context.Context, userId int64) ([]*dto.GetOrdersResponse, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.GetOrdersResponse, 0, len(orders))

	for _, order := range orders {
		res = append(res, &dto.GetOrdersResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	return res, nil
}
