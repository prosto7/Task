package model

import "time"

//easyjson:json
type Price struct {
	Symbol    string    `json:"symbol" gorm:"not null"`
	Price     string    `json:"price" gorm:"not null"`
	TimeStamp time.Time `gorm:""`
}
