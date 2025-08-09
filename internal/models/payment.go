package models

type Payment struct {
	ID           uint   `gorm:"primaryKey"`
	OrderUID     string `gorm:"unique;not null"`
	Transaction  string `json:"transaction" gorm:"not null"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" gorm:"not null"`
	Provider     string `json:"provider" gorm:"not null"`
	Amount       int    `json:"amount" gorm:"not null"`
	PaymentDT    int    `json:"payment_dt" gorm:"not null"`
	Bank         string `json:"bank" gorm:"not null"`
	DeliveryCost int    `json:"delivery_cost" gorm:"not null"`
	GoodsTotal   int    `json:"goods_total" gorm:"not null"`
	CustomFee    int    `json:"custom_fee" gorm:"not null"`
}
