package models

import "time"

type Item struct {
	Id        string     `json:"id"`
	ItemName  string     `json:"item_name"`
	SKU       string     `json:"sku"`
	Price     int64      `json:"Price"`
	Quantity  int        `json:"quantity"`
	Stock     int        `json:"stock"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
