package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cachemocks "github.com/fathersson/wb-demo-service/internal/cache/cachemocks"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
	repomocks "github.com/fathersson/wb-demo-service/internal/repository/repomocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestOrderHandler_CacheHit
// Проверяет сценарий, когда заказ найден в кэше
// GetCache("id1") возвращает order и true
// Репозиторий не вызывается
// Ожидаем ответ 200 и выполнение всех ожиданий моков
func TestOrderHandler_CacheHit(t *testing.T) {
	cache := cachemocks.NewCacheInterface(t)
	repo := repomocks.NewOrderRepository(t)
	order := models.Order{OrderUID: "id1"}

	cache.EXPECT().GetCache("id1").Return(order, true)

	srv := NewServer(config.HttpServer{Port: 8080}, cache, repo)
	req := httptest.NewRequest(http.MethodGet, "/order/id1", nil)
	w := httptest.NewRecorder()

	srv.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// TestOrderHandler_FromDB
// Проверяет поведение при отсутствии заказа в кэше
// 1) GetCache - miss
// 2) GetOrderById - возвращает order
// 3) SetCache вызывается для сохранения результата
// Ожидаем ответ 200
func TestOrderHandler_FromDB(t *testing.T) {
	cache := cachemocks.NewCacheInterface(t)
	repo := repomocks.NewOrderRepository(t)
	order := models.Order{OrderUID: "id2"}

	cache.EXPECT().GetCache("id2").Return(models.Order{}, false)
	repo.EXPECT().GetOrderById(mock.Anything, "id2").Return(order, nil)
	cache.EXPECT().SetCache("id2", order).Return()

	srv := NewServer(config.HttpServer{Port: 8080}, cache, repo)
	req := httptest.NewRequest(http.MethodGet, "/order/id2", nil)
	w := httptest.NewRecorder()

	srv.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// TestOrderHandler_NotFound
// Проверяет обработку случая, когда заказа нет ни в кэше, ни в БД
// Repo.GetOrderById возвращает ошибку
// Ожидаем 404 Not Found
func TestOrderHandler_NotFound(t *testing.T) {
	cache := cachemocks.NewCacheInterface(t)
	repo := repomocks.NewOrderRepository(t)

	cache.EXPECT().GetCache("missing").Return(models.Order{}, false)
	repo.EXPECT().GetOrderById(mock.Anything, "missing").Return(models.Order{}, assert.AnError)

	srv := NewServer(config.HttpServer{Port: 8080}, cache, repo)
	req := httptest.NewRequest(http.MethodGet, "/order/missing", nil)
	w := httptest.NewRecorder()

	srv.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// TestOrderHandler_MethodNotAllowed
// Проверяет, что POST-запрос к /order/{id} запрещён
// Ожидаем 405
// Кэш и репозиторий не должны вызываться
func TestOrderHandler_MethodNotAllowed(t *testing.T) {
	cache := cachemocks.NewCacheInterface(t)
	repo := repomocks.NewOrderRepository(t)

	srv := NewServer(config.HttpServer{Port: 8080}, cache, repo)
	req := httptest.NewRequest(http.MethodPost, "/order/id1", nil)
	w := httptest.NewRecorder()

	srv.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}
