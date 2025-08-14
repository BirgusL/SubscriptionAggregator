package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"SubscriptionAggregator/pkg/model"
	"SubscriptionAggregator/pkg/service"
)

type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) CreateSubscription(ctx context.Context, req service.CreateSubscriptionRequest) (*model.Subscription, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) UpdateSubscription(ctx context.Context, req service.UpdateSubscriptionRequest) (*model.Subscription, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionService) ListSubscriptions(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func newTestRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	return httptest.NewRequest(method, path, &buf)
}

func newTestHandler() (*SubscriptionHandler, *MockSubscriptionService) {
	mockSvc := &MockSubscriptionService{}
	return NewSubscriptionHandler(mockSvc), mockSvc
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dest); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
}

func TestCreateSubscription_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	reqBody := service.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   time.Now().Round(0),
	}

	expectedSub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: reqBody.ServiceName,
		Price:       reqBody.Price,
		UserID:      reqBody.UserID,
		StartDate:   reqBody.StartDate,
	}

	mockSvc.On("CreateSubscription", mock.Anything, reqBody).Return(expectedSub, nil)

	r := newTestRequest(http.MethodPost, "/subscriptions", reqBody)
	h.CreateSubscription(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Subscription
	parseResponse(t, w, &response)
	assert.Equal(t, *expectedSub, response)
	mockSvc.AssertExpectations(t)
}

func TestCreateSubscription_InvalidPayload(t *testing.T) {
	h, _ := newTestHandler()
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString("{invalid}"))
	h.CreateSubscription(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "invalid request payload", response["error"])
}

func TestCreateSubscription_ServiceError(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	reqBody := service.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   time.Now().Round(0),
	}

	mockSvc.On("CreateSubscription", mock.Anything, reqBody).Return(&model.Subscription{}, errors.New("db error"))

	r := newTestRequest(http.MethodPost, "/subscriptions", reqBody)
	h.CreateSubscription(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "db error", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestGetSubscription_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	subID := uuid.New()
	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       599,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   time.Now().Round(0),
	}

	mockSvc.On("GetSubscription", mock.Anything, subID).Return(expectedSub, nil)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := httptest.NewRequest(http.MethodGet, "/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var response model.Subscription
	parseResponse(t, w, &response)
	assert.Equal(t, *expectedSub, response)
	mockSvc.AssertExpectations(t)
}

func TestGetSubscription_InvalidID(t *testing.T) {
	h, _ := newTestHandler()
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := httptest.NewRequest(http.MethodGet, "/subscriptions/invalid", nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "invalid subscription ID", response["error"])
}

func TestGetSubscription_NotFound(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	subID := uuid.New()
	mockSvc.On("GetSubscription", mock.Anything, subID).Return(&model.Subscription{}, model.ErrNotFound)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := httptest.NewRequest(http.MethodGet, "/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "subscription not found", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateSubscription_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	subID := uuid.New()
	reqBody := service.UpdateSubscriptionRequest{
		ServiceName: "Yandex Plus Premium",
		Price:       799,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   time.Now().Round(0),
	}

	expectedSub := &model.Subscription{
		ID:          subID,
		ServiceName: reqBody.ServiceName,
		Price:       reqBody.Price,
		UserID:      reqBody.UserID,
		StartDate:   reqBody.StartDate,
	}

	mockSvc.On("UpdateSubscription", mock.Anything, mock.MatchedBy(func(req service.UpdateSubscriptionRequest) bool {
		return req.ID == subID
	})).Return(expectedSub, nil)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := newTestRequest(http.MethodPut, "/subscriptions/"+subID.String(), reqBody)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var response model.Subscription
	parseResponse(t, w, &response)
	assert.Equal(t, *expectedSub, response)
	mockSvc.AssertExpectations(t)
}

func TestUpdateSubscription_IDMismatch(t *testing.T) {
	h, _ := newTestHandler()
	w := httptest.NewRecorder()

	reqBody := service.UpdateSubscriptionRequest{
		ServiceName: "Yandex Plus Premium",
		Price:       799,
	}

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := newTestRequest(http.MethodPut, "/subscriptions/invalid", reqBody)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "invalid subscription ID", response["error"])
}

func TestDeleteSubscription_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	subID := uuid.New()
	mockSvc.On("DeleteSubscription", mock.Anything, subID).Return(nil)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 5, w.Body.Len())
	mockSvc.AssertExpectations(t)
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	subID := uuid.New()
	mockSvc.On("DeleteSubscription", mock.Anything, subID).Return(model.ErrNotFound)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	r := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	parseResponse(t, w, &response)
	assert.Equal(t, "subscription not found", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestListSubscriptions_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	fromDate := time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 8, 12, 0, 0, 0, 0, time.UTC)

	expectedSubs := []*model.Subscription{
		{
			ID:          uuid.New(),
			ServiceName: "Yandex Plus",
			Price:       599,
			UserID:      userID,
			StartDate:   time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	mockSvc.On("ListSubscriptions", mock.Anything, mock.MatchedBy(func(filter model.SubscriptionFilter) bool {
		return filter.UserID != nil && *filter.UserID == userID &&
			filter.FromDate != nil && filter.FromDate.Equal(fromDate) &&
			filter.ToDate != nil && filter.ToDate.Equal(toDate)
	})).Return(expectedSubs, nil)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	url := fmt.Sprintf("/subscriptions?user_id=%s&from_date=%s&to_date=%s",
		userID.String(),
		fromDate.Format(time.RFC3339),
		toDate.Format(time.RFC3339))

	r := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []*model.Subscription
	parseResponse(t, w, &response)
	assert.Equal(t, expectedSubs, response)
	mockSvc.AssertExpectations(t)
}

func TestGetTotalCost_Success(t *testing.T) {
	h, mockSvc := newTestHandler()
	w := httptest.NewRecorder()

	serviceName := "Yandex Plus"
	expectedTotal := 1500

	mockSvc.On("GetTotalCost", mock.Anything, mock.MatchedBy(func(f model.SubscriptionFilter) bool {
		return f.ServiceName != nil && *f.ServiceName == serviceName
	})).Return(expectedTotal, nil)

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	url := "/subscriptions/total?service_name=" + url.QueryEscape(serviceName)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]int
	parseResponse(t, w, &response)
	assert.Equal(t, expectedTotal, response["total"])
	mockSvc.AssertExpectations(t)
}
