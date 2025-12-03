package models

import "fmt"

func ValidateOrder(o Order) error {
	// 1. Обязательные поля верхнего уровня
	if o.OrderUID == "" {
		return fmt.Errorf("order_uid is empty")
	}
	if o.TrackNumber == "" {
		return fmt.Errorf("track_number is empty")
	}
	if o.CustomerID == "" {
		return fmt.Errorf("customer_id is empty")
	}
	if o.DeliveryService == "" {
		return fmt.Errorf("delivery_service is empty")
	}
	if o.Locale == "" {
		return fmt.Errorf("locale is empty")
	}

	// 2. Delivery
	if o.Delivery.Name == "" {
		return fmt.Errorf("delivery.name is empty")
	}
	if o.Delivery.Phone == "" {
		return fmt.Errorf("delivery.phone is empty")
	}
	if o.Delivery.City == "" {
		return fmt.Errorf("delivery.city is empty")
	}
	if o.Delivery.Address == "" {
		return fmt.Errorf("delivery.address is empty")
	}

	// 3. Payment
	if o.Payment.Transaction == "" {
		return fmt.Errorf("payment.transaction is empty")
	}
	// Часто transaction должен совпадать с order_uid
	if o.Payment.Transaction != o.OrderUID {
		return fmt.Errorf("payment.transaction != order_uid")
	}
	if o.Payment.Currency == "" {
		return fmt.Errorf("payment.currency is empty")
	}
	if o.Payment.Provider == "" {
		return fmt.Errorf("payment.provider is empty")
	}
	if o.Payment.Amount <= 0 {
		return fmt.Errorf("payment.amount must be > 0")
	}
	if o.Payment.DeliveryCost < 0 {
		return fmt.Errorf("payment.delivery_cost must be >= 0")
	}
	if o.Payment.GoodsTotal < 0 {
		return fmt.Errorf("payment.goods_total must be >= 0")
	}
	if o.Payment.PaymentDT <= 0 {
		return fmt.Errorf("payment.payment_dt must be > 0")
	}

	// 4. Items
	if len(o.Items) == 0 {
		return fmt.Errorf("items list is empty")
	}
	for i, it := range o.Items {
		if it.ChrtID == 0 {
			return fmt.Errorf("items[%d].chrt_id is zero", i)
		}
		if it.Name == "" {
			return fmt.Errorf("items[%d].name is empty", i)
		}
		if it.Price < 0 {
			return fmt.Errorf("items[%d].price must be >= 0", i)
		}
		if it.TotalPrice < 0 {
			return fmt.Errorf("items[%d].total_price must be >= 0", i)
		}
		if it.TrackNumber != "" && it.TrackNumber != o.TrackNumber {
			return fmt.Errorf("items[%d].track_number != order.track_number", i)
		}
	}

	return nil
}
