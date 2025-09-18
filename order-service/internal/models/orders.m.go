package models

import "time"

type Order struct {
	ID         string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProductID  string    `gorm:"column:productId;type:uuid;not null" json:"productId"`
	TotalPrice float64   `gorm:"column:totalPrice;not null" json:"totalPrice"`
	Status     string    `gorm:"not null" json:"status"`
	CreatedAt  time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}
