package main

import (
	"fmt"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func newDB(cfg *Config, pass string) (*gorm.DB, error) {
	db, err := gorm.Open(dialectDB, dbConnString(cfg, pass))
	if err != nil {
		return nil, err
	}
	if err := db.DB().Ping(); err != nil {
		return nil, err
	}
	// db.LogMode(true)
	return db, nil
}

func dbConnString(cfg *Config, pass string) (cns string) {
	var sslmode string
	switch cfg.ConnectDB.SslMode {
	case true:
		sslmode = "enable"
	case false:
		sslmode = "disable"
	}
	switch dialectDB {
	case "mysql":
		cns = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=%s",
			cfg.ConnectDB.User,
			pass,
			cfg.ConnectDB.Database,
			cfg.ConnectDB.Host+":"+cfg.ConnectDB.Port)
	case "postgres":
		cns = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			cfg.ConnectDB.Host,
			cfg.ConnectDB.Port,
			cfg.ConnectDB.User,
			cfg.ConnectDB.Database,
			pass,
			sslmode)
	case "sqlite3":
		cns = cfg.ConnectDB.SQLitePath
	}
	return
}
