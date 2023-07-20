package models

import "time"

type Order struct {
	Id                   string      `json:"id"`
	UserId               string      `json:"user_id"`
	Status               int         `json:"status"`
	OrderItem            []OrderItem `json:"order_item"`
	TotalAmount          int64       `json:"total_amount"`
	TotalQuantity        int         `json:"total_quantity"`
	TotalDiscountAmount  int64       `json:"total_discount_amount"`
	PaymentMethod        string      `json:"payment_method"`
	PaymentAcquirementId string      `json:"payment_acquirement_id"`
	PaymentDate          *time.Time  `json:"payment_date"`
	CreatedAt            *time.Time  `json:"created_at"`
	UpdatedAt            *time.Time  `json:"updated_at"`
}

type OrderItem struct {
	Id             string     `json:"id"`
	OrderId        string     `json:"order_id"`
	ItemId         string     `json:"item_id"`
	ItemName       string     `json:"item_name"`
	SKU            string     `json:"sku"`
	Quantity       int        `json:"quantity"`
	ItemPrice      int64      `json:"item_price"`
	DiscountAmount int64      `json:"discount_amount"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

type InsertOrder struct {
	Id                   string            `json:"id"`
	UserId               string            `json:"user_id"`
	Status               int               `json:"status"`
	OrderItem            []InsertOrderItem `gorm:"-" json:"order_item"`
	TotalAmount          int64             `json:"total_amount"`
	TotalQuantity        int               `json:"total_quantity"`
	TotalDiscountAmount  int64             `json:"total_discount_amount"`
	PaymentMethod        string            `json:"payment_method"`
	PaymentAcquirementId string            `json:"payment_acquirement_id"`
	PaymentDate          *time.Time        `json:"payment_date"`
	CreatedAt            *time.Time        `json:"created_at"`
	UpdatedAt            *time.Time        `json:"updated_at"`
}

type InsertOrderItem struct {
	Id             string     `json:"id"`
	OrderId        string     `json:"order_id"`
	ItemId         string     `json:"item_id"`
	Quantity       int        `json:"quantity"`
	ItemPrice      int64      `json:"item_price"`
	DiscountAmount int64      `json:"discount_amount"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

type CreateOrder struct {
	UserId    string `json:"user_id" binding:"required"`
	OrderItem []struct {
		ItemId   string `json:"item_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	} `json:"order_item"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

type UpdateOrder struct {
	OrderId   string `json:"order_id" binding:"required"`
	UserId    string `json:"user_id" binding:"required"`
	OrderItem []struct {
		ItemId   string `json:"item_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	} `json:"order_item"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

type OrderLog struct {
	OrderId     string     `json:"order_id"`
	OrderStatus int        `json:"order_status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
