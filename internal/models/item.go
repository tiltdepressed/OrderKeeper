package models

type Item struct {
	ID          uint   `gorm:"primaryKey"`
	OrderUID    string `gorm:"index;not null"`
	CHRTID      int    `json:"chrt_id" gorm:"not null"`
	TrackNumber string `json:"track_number" gorm:"not null"`
	Price       int    `json:"price" gorm:"not null"`
	RID         string `json:"rid" gorm:"not null"`
	Name        string `json:"name" gorm:"not null"`
	Sale        int    `json:"sale" gorm:"not null"`
	Size        string `json:"size" gorm:"not null"`
	TotalPrice  int    `json:"total_price" gorm:"not null"`
	NMID        int    `json:"nm_id" gorm:"not null"`
	Brand       string `json:"brand" gorm:"not null"`
	Status      int    `json:"status" gorm:"not null"`
}
