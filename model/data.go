package model

import (
	"sync"
	"time"
)

type Data struct {
	TimeStamp time.Time
	Prices    []Price
	Wg        *sync.WaitGroup
}
