package models

import (
	"time"
)

// Order — корневая структура заказа
type Order struct {
	OrderUID          string   `json:"order_uid" validate:"required,alphanum"` // Буквы и цифры`
	TrackNumber       string   `json:"track_number" validate:"required"`       // Не пустая строка`
	Entry             string   `json:"entry,omitempty" validate:"omitempty"`   // Не пустая строка`
	Delivery          Delivery `json:"delivery" validate:"required,dive"`
	Payment           Payment  `json:"payment" validate:"required,dive"`
	Items             []Item   `json:"items" validate:"required,dive"`
	Locale            string   `json:"locale,omitempty" validate:"omitempty"` // Только буквы``
	InternalSignature string   `json:"internal_signature,omitempty"`
	CustomerID        string   `json:"customer_id,omitempty" validate:"omitempty"`
	DeliveryService   string   `json:"delivery_service,omitempty" validate:"omitempty"`
	ShardKey          string   `json:"shardkey,omitempty" validate:"omitempty,numeric"`
	SmID              int      `json:"sm_id,omitempty" validate:"omitempty,min=0"`
	// DateCreated: если в JSON ISO-8601 (например "2023-10-01T12:34:56Z"),
	// используем time.Time. Если приходит unix-timestamp — используйте поле Payment.PaymentDT.
	DateCreated time.Time `json:"date_created,omitempty" validate:"omitempty"`
	OofShard    string    `json:"oof_shard,omitempty" validate:"omitempty,numeric"`
}

// Delivery — данные доставки
type Delivery struct {
	Name    string `json:"name" validate:"required,alphaunicode"` // Только буквы`
	Phone   string `json:"phone" validate:"required,e164"`
	Zip     string `json:"zip" validate:"required,numeric"`
	City    string `json:"city" validate:"required,alphaunicode"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region,omitempty" validate:"omitempty"`
	Email   string `json:"email,omitempty" validate:"omitempty,email"`
}

// Payment — платёжные данные
type Payment struct {
	Transaction  string `json:"transaction" validate:"required,alphanum"`
	RequestID    string `json:"request_id,omitempty"`
	Currency     string `json:"currency" validate:"required"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,min=0"`     // сумма в целых единицах (например, копейки/центы или просто рубли — договоритесь внутри команды)
	PaymentDT    int64  `json:"payment_dt" validate:"required,min=0"` // unix timestamp (seconds). Альтернатива — ISO date -> use Order.DateCreated
	Bank         string `json:"bank,omitempty" validate:"omitempty"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,min=0"` // стоимость доставки
	GoodsTotal   int    `json:"goods_total" validate:"required,min=0"`   // сумма товаров
	CustomFee    int    `json:"custom_fee,omitempty" validate:"omitempty,min=0"`
}

// Item — один товар в заказе
type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,min=0"`           // id в каталоге
	TrackNumber string `json:"track_number,omitempty" validate:"omitempty"` // трек-номер позиции (если есть)
	Price       int    `json:"price" validate:"required,min=0"`             // цена за единицу
	Rid         string `json:"rid,omitempty" validate:"omitempty,alphanum"` // request id / внутренний id
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale,omitempty"` // скидка в процентах или в абсолюте — договоритесь
	Size        string `json:"size,omitempty"`
	TotalPrice  int    `json:"total_price,omitempty" validate:"omitempty,min=0"` // цена * количество (если в JSON есть)
	NmID        int    `json:"nm_id,omitempty" validate:"omitempty,min=0"`       // ещё один id
	Brand       string `json:"brand,omitempty" validate:"omitempty"`
	Status      int    `json:"status,omitempty" validate:"omitempty,min=0"`
}
