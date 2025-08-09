// Package models
package models

type Order struct {
	OrderUID          string   `json:"order_uid" gorm:"primaryKey;unique;not null"`
	TrackNumber       string   `json:"track_number" gorm:"not null"`
	Entry             string   `json:"entry" gorm:"not null"`
	Delivery          Delivery `json:"delivery" gorm:"foreignKey:OrderUID;constraint:OnDelete:CASCADE;"`
	Payment           Payment  `json:"payment" gorm:"foreignKey:OrderUID;constraint:OnDelete:CASCADE;"`
	Items             []Item   `json:"items" gorm:"foreignKey:OrderUID;constraint:OnDelete:CASCADE;"`
	Locale            string   `json:"locale" gorm:"not null"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id" gorm:"not null"`
	DeliveryService   string   `json:"delivery_service" gorm:"not null"`
	Shardkey          string   `json:"shardkey" gorm:"not null"`
	SMID              int      `json:"sm_id" gorm:"not null"`
	DateCreated       string   `json:"date_created" gorm:"not null"`
	OOFShard          string   `json:"oof_shard" gorm:"not null"`
}
