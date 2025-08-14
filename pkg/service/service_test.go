package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"SubscriptionAggregator/pkg/model"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) List(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func newTestService() (*subscriptionService, *MockSubscriptionRepository) {
	mockRepo := &MockSubscriptionRepository{}
	return NewSubscriptionService(mockRepo).(*subscriptionService), mockRepo
}

func fixedTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

func fixedUUID() uuid.UUID {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
}

func TestCreateSubscription_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	req := CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      fixedUUID(),
		StartDate:   fixedTime(),
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(sub *model.Subscription) bool {
		return sub.ServiceName == req.ServiceName &&
			sub.Price == req.Price &&
			sub.UserID == req.UserID &&
			sub.StartDate.Equal(req.StartDate)
	})).Return(nil)

	sub, err := s.CreateSubscription(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, req.ServiceName, sub.ServiceName)
	assert.Equal(t, req.Price, sub.Price)
	assert.Equal(t, req.UserID, sub.UserID)
	assert.Equal(t, req.StartDate, sub.StartDate)
	mockRepo.AssertExpectations(t)
}

func TestCreateSubscription_RepositoryError(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	req := CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      fixedUUID(),
		StartDate:   fixedTime(),
	}

	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

	sub, err := s.CreateSubscription(ctx, req)

	assert.Nil(t, sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create subscription")
	mockRepo.AssertExpectations(t)
}

func TestGetSubscription_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()
	subID := fixedUUID()

	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      fixedUUID(),
		StartDate:   fixedTime(),
	}

	mockRepo.On("GetByID", ctx, subID).Return(expectedSub, nil)

	sub, err := s.GetSubscription(ctx, subID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSub, sub)
	mockRepo.AssertExpectations(t)
}

func TestGetSubscription_NotFound(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()
	subID := fixedUUID()

	mockRepo.On("GetByID", ctx, subID).Return((*model.Subscription)(nil), model.ErrNotFound)

	sub, err := s.GetSubscription(ctx, subID)

	assert.Nil(t, sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get subscription")
	mockRepo.AssertExpectations(t)
}

func TestUpdateSubscription_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	req := UpdateSubscriptionRequest{
		ID:          fixedUUID(),
		ServiceName: "Yandex Plus",
		Price:       799,
		UserID:      fixedUUID(),
		StartDate:   fixedTime(),
	}

	expectedSub := &model.Subscription{
		ID:          req.ID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
	}

	mockRepo.On("Update", ctx, expectedSub).Return(nil)

	sub, err := s.UpdateSubscription(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedSub, sub)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSubscription_RepositoryError(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	req := UpdateSubscriptionRequest{
		ID:          fixedUUID(),
		ServiceName: "Yandex Plus",
		Price:       799,
		UserID:      fixedUUID(),
		StartDate:   fixedTime(),
	}

	mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	sub, err := s.UpdateSubscription(ctx, req)

	assert.Nil(t, sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update subscription")
	mockRepo.AssertExpectations(t)
}

func TestDeleteSubscription_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()
	subID := fixedUUID()

	mockRepo.On("Delete", ctx, subID).Return(nil)

	err := s.DeleteSubscription(ctx, subID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteSubscription_RepositoryError(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()
	subID := fixedUUID()

	mockRepo.On("Delete", ctx, subID).Return(errors.New("db error"))

	err := s.DeleteSubscription(ctx, subID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete subscription")
	mockRepo.AssertExpectations(t)
}

func TestListSubscriptions_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	filter := model.SubscriptionFilter{
		UserID:      &[]uuid.UUID{fixedUUID()}[0],
		ServiceName: &[]string{"Yandex Plus"}[0],
	}

	expectedSubs := []*model.Subscription{
		{
			ID:          fixedUUID(),
			ServiceName: "Yandex Plus",
			Price:       599,
			UserID:      fixedUUID(),
			StartDate:   fixedTime(),
		},
	}

	mockRepo.On("List", ctx, filter).Return(expectedSubs, nil)

	subs, err := s.ListSubscriptions(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedSubs, subs)
	mockRepo.AssertExpectations(t)
}

func TestListSubscriptions_RepositoryError(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	filter := model.SubscriptionFilter{
		UserID: &[]uuid.UUID{fixedUUID()}[0],
	}

	mockRepo.On("List", ctx, filter).Return([]*model.Subscription(nil), errors.New("db error"))

	subs, err := s.ListSubscriptions(ctx, filter)

	assert.Nil(t, subs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list subscriptions")
	mockRepo.AssertExpectations(t)
}

func TestGetTotalCost_Success(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	filter := model.SubscriptionFilter{
		ServiceName: &[]string{"Yandex Plus"}[0],
	}
	expectedTotal := 1500

	mockRepo.On("GetTotalCost", ctx, filter).Return(expectedTotal, nil)

	total, err := s.GetTotalCost(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

func TestGetTotalCost_RepositoryError(t *testing.T) {
	s, mockRepo := newTestService()
	ctx := context.Background()

	filter := model.SubscriptionFilter{
		ServiceName: &[]string{"Yandex Plus"}[0],
	}

	mockRepo.On("GetTotalCost", ctx, filter).Return(0, errors.New("db error"))

	total, err := s.GetTotalCost(ctx, filter)

	assert.Equal(t, 0, total)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to calculate total cost")
	mockRepo.AssertExpectations(t)
}
