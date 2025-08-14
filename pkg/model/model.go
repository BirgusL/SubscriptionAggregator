package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceName string     `json:"service_name" example:"Yandex Plus"`
	Price       int        `json:"price" example:"599"`
	UserID      uuid.UUID  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   time.Time  `json:"start_date" example:"2025-08-12T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2025-09-12T00:00:00Z"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string    `json:"service_name" example:"Yandex Plus"`
	FromDate    *time.Time `json:"from_date" example:"2025-08-12T00:00:00Z"`
	ToDate      *time.Time `json:"to_date" example:"2025-09-12T00:00:00Z"`
}

// Custom errors for handlers
var (
	ErrNotFound = errors.New("not found")
)

// ***
// Custom responses for swagger
type ErrorResponse struct {
	Error string `json:"error" example:"invalid subscription ID"`
	Code  int    `json:"code" example:"404"`
}

type ErrorInput struct {
	Error string `json:"error" example:"invalid input"`
	Code  int    `json:"code" example:"400"`
}

type TotalCostResponse struct {
	Total int `json:"total" example:"1500"`
}

type SubscriptionListResponse struct {
	Subscriptions []*Subscription `json:"subscriptions"`
	Count         int             `json:"count" example:"5"`
}

type ServerError struct {
	Error string `json:"error" example:"server does not respond"`
	Code  int    `json:"code" example:"500"`
}

//***
