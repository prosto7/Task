package model

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

type DB struct {
	Db  *gorm.DB
	Err error
}

func (db *DB) ConnectDB(dsn string) {
	db.Db, db.Err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: glog.Default.LogMode(glog.Info),
	})
}
