package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"SubscriptionAggregator/pkg/model"
	"SubscriptionAggregator/pkg/service"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
}

func NewSubscriptionHandler(service service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/subscriptions", h.CreateSubscription).Methods("POST")
	router.HandleFunc("/subscriptions/total", h.GetTotalCost).Methods("GET")
	router.HandleFunc("/subscriptions/{id}", h.GetSubscription).Methods("GET")
	router.HandleFunc("/subscriptions/{id}", h.UpdateSubscription).Methods("PUT")
	router.HandleFunc("/subscriptions/{id}", h.DeleteSubscription).Methods("DELETE")
	router.HandleFunc("/subscriptions", h.ListSubscriptions).Methods("GET")
}

// CreateSubscription создает новую подписку
// @Summary Создать подписку
// @Description Добавляет новую подписку для пользователя
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param input body service.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} model.Subscription "Подписка успешно создана"
// @SuccessExample {json} Success-Response:
//     HTTP/1.1 201 Created
//     {
//         "id": "550e8400-e29b-41d4-a716-446655440000",
//         "service_name": "Yandex Plus",
//         "price": 599,
//         "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
//         "start_date": "2025-01-01T00:00:00Z"
//     }
// @Failure 400 {object} model.ErrorInput "Неверный формат данных"
// @FailureExample {json} Error-Response:
//     HTTP/1.1 400 Bad Request
//     {
//         "error": "invalid request payload",
//         "code": 400
//     }
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions [post]

func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	sub, err := h.service.CreateSubscription(r.Context(), req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, sub)
}

// GetSubscription возвращает подписку по ID
// @Summary Получить подписку
// @Description Возвращает информацию о конкретной подписке
// @Tags Subscriptions
// @Produce json
// @Param id path string true "ID подписки" example(550e8400-e29b-41d4-a716-446655440000)
// @Success 200 {object} model.Subscription
// @SuccessExample {json} Success-Response:
//
//	HTTP/1.1 200 OK
//	{
//	    "id": "550e8400-e29b-41d4-a716-446655440000",
//	    "service_name": "Yandex Plus",
//	    "price": 599,
//	    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
//	    "start_date": "2025-01-01T00:00:00Z"
//	}
//
// @Failure 400 {object} model.ErrorInput
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 400 Bad Request
//	{
//	    "error": "invalid subscription ID",
//	    "code": 400
//	}
//
// @Failure 404 {object} model.ErrorResponse
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 404 Not Found
//	{
//	    "error": "subscription not found",
//	    "code": 404
//	}
//
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	sub, err := h.service.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

// UpdateSubscription обновляет существующую подписку
// @Summary Обновить подписку
// @Description Изменяет данные существующей подписки
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки" example(550e8400-e29b-41d4-a716-446655440000)
// @Param input body service.UpdateSubscriptionRequest true "Новые данные подписки"
// @Success 200 {object} model.Subscription "Подписка успешно обновлена"
// @SuccessExample {json} Success-Response:
//
//	HTTP/1.1 200 OK
//	{
//	    "id": "550e8400-e29b-41d4-a716-446655440000",
//	    "service_name": "Yandex Plus",
//	    "price": 799,
//	    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
//	    "start_date": "2025-01-01T00:00:00Z",
//	    "end_date": "2025-12-31T00:00:00Z"
//	}
//
// @Failure 400 {object} model.ErrorInput "Неверный формат данных"
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 400 Bad Request
//	{
//	    "error": "invalid request payload",
//	    "code": 400
//	}
//
// @Failure 404 {object} model.ErrorResponse "Подписка не найдена"
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 404 Not Found
//	{
//	    "error": "subscription not found",
//	    "code": 404
//	}
//
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	var req service.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	req.ID = id

	sub, err := h.service.UpdateSubscription(r.Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

// DeleteSubscription удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Tags Subscriptions
// @Param id path string true "ID подписки" example(550e8400-e29b-41d4-a716-446655440000)
// @Success 204 "Подписка успешно удалена"
// @Failure 400 {object} model.ErrorInput "Неверный ID подписки"
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 400 Bad Request
//	{
//	    "error": "invalid subscription ID",
//	    "code": 400
//	}
//
// @Failure 404 {object} model.ErrorResponse "Подписка не найдена"
// @FailureExample {json} Error-Response:
//
//	HTTP/1.1 404 Not Found
//	{
//	    "error": "subscription not found",
//	    "code": 404
//	}
//
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	if err := h.service.DeleteSubscription(r.Context(), id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// ListSubscriptions возвращает список подписок с фильтрацией
// @Summary Список подписок
// @Description Возвращает подписки с возможностью фильтрации
// @Tags Subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя" example(60601fee-2bf1-4721-ae6f-7636e79a0cba)
// @Param service_name query string false "Название сервиса" example(Yandex Plus)
// @Param from_date query string false "Начальная дата (RFC3339)" example(2025-01-01T00:00:00Z)
// @Param to_date query string false "Конечная дата (RFC3339)" example(2025-12-31T00:00:00Z)
// @Success 200 {array} model.Subscription
// @SuccessExample {json} Success-Response:
//
//	HTTP/1.1 200 OK
//	[
//	    {
//	        "id": "550e8400-e29b-41d4-a716-446655440000",
//	        "service_name": "Yandex Plus",
//	        "price": 599,
//	        "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
//	        "start_date": "2025-01-01T00:00:00Z"
//	    }
//	]
//
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	filter := model.SubscriptionFilter{
		UserID:      getUUIDQueryParam(r, "user_id"),
		ServiceName: getStringQueryParam(r, "service_name"),
		FromDate:    getTimeQueryParam(r, "from_date"),
		ToDate:      getTimeQueryParam(r, "to_date"),
	}

	subs, err := h.service.ListSubscriptions(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, subs)
}

// GetTotalCost возвращает суммарную стоимость подписок
// @Summary Сумма подписок
// @Description Возвращает общую стоимость подписок за период
// @Tags Subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя" example(60601fee-2bf1-4721-ae6f-7636e79a0cba)
// @Param service_name query string false "Название сервиса" example(Yandex Plus)
// @Param from_date query string false "Начальная дата (RFC3339)" example(2025-01-01T00:00:00Z)
// @Param to_date query string false "Конечная дата (RFC3339)" example(2025-12-31T00:00:00Z)
// @Success 200 {object} model.TotalCostResponse
// @SuccessExample {json} Success-Response:
//
//	HTTP/1.1 200 OK
//	{
//	    "total": 1500
//	}
//
// @Failure 500 {object} model.ServerError "Ошибка сервера"
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	filter := model.SubscriptionFilter{
		UserID:      getUUIDQueryParam(r, "user_id"),
		ServiceName: getStringQueryParam(r, "service_name"),
		FromDate:    getTimeQueryParam(r, "from_date"),
		ToDate:      getTimeQueryParam(r, "to_date"),
	}

	total, err := h.service.GetTotalCost(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int{"total": total})
}

// ***
// Helper funcs
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func getUUIDQueryParam(r *http.Request, param string) *uuid.UUID {
	val := r.URL.Query().Get(param)
	if val == "" {
		return nil
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return nil
	}
	return &id
}

func getStringQueryParam(r *http.Request, param string) *string {
	val := r.URL.Query().Get(param)
	if val == "" {
		return nil
	}
	return &val
}

func getTimeQueryParam(r *http.Request, param string) *time.Time {
	val := r.URL.Query().Get(param)
	if val == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil
	}
	return &t
}

//***
