package models

type Delivery struct {
	ID       uint   `gorm:"primaryKey"`
	OrderUID string `gorm:"unique;not null"`
	Name     string `json:"name" gorm:"not null"`
	Phone    string `json:"phone" gorm:"not null"`
	Zip      string `json:"zip" gorm:"not null"`
	City     string `json:"city" gorm:"not null"`
	Address  string `json:"address" gorm:"not null"`
	Region   string `json:"region" gorm:"not null"`
	Email    string `json:"email" gorm:"not null"`
}
