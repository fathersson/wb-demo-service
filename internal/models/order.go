package models

import "time"

// Order — корневая структура заказа
type Order struct {
	OrderUID          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry,omitempty"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale,omitempty"`
	InternalSignature string   `json:"internal_signature,omitempty"`
	CustomerID        string   `json:"customer_id,omitempty"`
	DeliveryService   string   `json:"delivery_service,omitempty"`
	ShardKey          string   `json:"shardkey,omitempty"`
	SmID              int      `json:"sm_id,omitempty"`
	// DateCreated: если в JSON ISO-8601 (например "2023-10-01T12:34:56Z"),
	// используем time.Time. Если приходит unix-timestamp — используйте поле Payment.PaymentDT.
	DateCreated time.Time `json:"date_created,omitempty"`
	// OofShard и т.п. — опциональные поля
	OofShard string `json:"oof_shard,omitempty"`
}

// Delivery — данные доставки
type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region,omitempty"`
	Email   string `json:"email,omitempty"`
}

// Payment — платёжные данные
type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id,omitempty"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`     // сумма в целых единицах (например, копейки/центы или просто рубли — договоритесь внутри команды)
	PaymentDT    int64  `json:"payment_dt"` // unix timestamp (seconds). Альтернатива — ISO date -> use Order.DateCreated
	Bank         string `json:"bank,omitempty"`
	DeliveryCost int    `json:"delivery_cost"` // стоимость доставки
	GoodsTotal   int    `json:"goods_total"`   // сумма товаров
	CustomFee    int    `json:"custom_fee,omitempty"`
}

// Item — один товар в заказе
type Item struct {
	ChrtID      int    `json:"chrt_id"`                // id в каталоге
	TrackNumber string `json:"track_number,omitempty"` // трек-номер позиции (если есть)
	Price       int    `json:"price"`                  // цена за единицу
	Rid         string `json:"rid,omitempty"`          // request id / внутренний id
	Name        string `json:"name"`
	Sale        int    `json:"sale,omitempty"` // скидка в процентах или в абсолюте — договоритесь
	Size        string `json:"size,omitempty"`
	TotalPrice  int    `json:"total_price,omitempty"` // цена * количество (если в JSON есть)
	NmID        int    `json:"nm_id,omitempty"`       // ещё один id
	Brand       string `json:"brand,omitempty"`
	Status      int    `json:"status,omitempty"`
	// Если нужен count/quantity, добавь поле:
	// Count int `json:"count,omitempty"`
}
