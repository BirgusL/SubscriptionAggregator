package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"SubscriptionAggregator/pkg/model"
	"SubscriptionAggregator/pkg/repository"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, req UpdateSubscriptionRequest) (*model.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error)
	GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error)
}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

type CreateSubscriptionRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*model.Subscription, error) {
	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

type UpdateSubscriptionRequest struct {
	ID          uuid.UUID  `json:"-"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, req UpdateSubscriptionRequest) (*model.Subscription, error) {
	sub := &model.Subscription{
		ID:          req.ID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error) {
	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

func (s *subscriptionService) GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	total, err := s.repo.GetTotalCost(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}
	return total, nil
}
