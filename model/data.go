package model

import "time"

type Data struct {
	TimeStamp time.Time
	Prices    []Price
}
