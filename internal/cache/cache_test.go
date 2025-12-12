package cache

import (
	"testing"

	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestSetCache_GetCache проверяет базовый сценарий:
// 1) Добавляем заказ в кэш через SetCache
// 2) Достаём его через GetCache
// 3) Убеждаемся, что ключ найден и данные совпадают
func TestSetCache_GetCache(t *testing.T) {
	c := NewCache()
	order := models.Order{OrderUID: "test123"}

	c.SetCache("test123", order)

	got, ok := c.GetCache("test123")
	assert.True(t, ok)
	assert.Equal(t, order.OrderUID, got.OrderUID)
}

// TestGetCache_NotFound убеждается, что запрос неизвестного ключа
// возвращает (пустое значение, false) и не паникует
func TestGetCache_NotFound(t *testing.T) {
	c := NewCache()

	_, ok := c.GetCache("missing")
	assert.False(t, ok)
}

// TestSetCache_Eviction проверяет механизм замены элементов при переполнении:
// 1) Ставим maxLen = 2 (маленький лимит для теста)
// 2) Кладём 3 заказа подряд
// 3) Самый старый (первый) должен быть удалён, последние два - остаться
func TestSetCache_Eviction(t *testing.T) {
	c := NewCache()
	c.maxLen = 2 // Устанавливаем маленький лимит для теста

	// Добавляем первый заказ
	order1 := models.Order{OrderUID: "order1"}
	c.SetCache("order1", order1)

	// Добавляем второй заказ
	order2 := models.Order{OrderUID: "order2"}
	c.SetCache("order2", order2)

	// Проверяем, что оба есть
	_, ok1 := c.GetCache("order1")
	_, ok2 := c.GetCache("order2")
	assert.True(t, ok1)
	assert.True(t, ok2)

	// Добавляем третий - должен вытеснить первый
	order3 := models.Order{OrderUID: "order3"}
	c.SetCache("order3", order3)

	// Первый должен быть удалён
	_, ok1 = c.GetCache("order1")
	assert.False(t, ok1, "order1 должен быть вытеснен")

	// Второй и третий должны остаться
	_, ok2 = c.GetCache("order2")
	_, ok3 := c.GetCache("order3")
	assert.True(t, ok2)
	assert.True(t, ok3)
}

// TestSetCache_Update проверяет обновление существующего ключа:
// 1) Кладём заказ с ключом X
// 2) Кладём новый заказ с тем же ключом X (другие поля)
// 3) При чтении должен вернуться обновлённый вариант
func TestSetCache_Update(t *testing.T) {
	c := NewCache()
	order1 := models.Order{OrderUID: "test123", TrackNumber: "OLD"}
	c.SetCache("test123", order1)

	order2 := models.Order{OrderUID: "test123", TrackNumber: "NEW"}
	c.SetCache("test123", order2)

	got, ok := c.GetCache("test123")
	assert.True(t, ok)
	assert.Equal(t, "NEW", got.TrackNumber)
}

// TestSetCache_Concurrent базово проверяет потокобезопасность:
// 1) В нескольких горутинах добавляем разные ключи
// 2) Ждём завершения всех горутин
// 3) Проверяем, что все добавленные ключи читаются без гонок
func TestSetCache_Concurrent(t *testing.T) {
	c := NewCache()
	done := make(chan bool)

	// Параллельно добавляем заказы
	for i := 0; i < 10; i++ {
		go func(id int) {
			order := models.Order{OrderUID: string(rune(id))}
			c.SetCache(string(rune(id)), order)
			done <- true
		}(i)
	}

	// Ждём завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что все заказы добавлены
	for i := 0; i < 10; i++ {
		_, ok := c.GetCache(string(rune(i)))
		assert.True(t, ok)
	}
}
