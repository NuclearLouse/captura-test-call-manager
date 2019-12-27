package database

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"redits.oculeus.com/asorokin/tcm/config"
)

type DB struct {
	Connect *gorm.DB
}

func NewDB(cfg *config.Config, pass string) (*DB, error) {
	db, err := gorm.Open(os.Getenv("DIALECT_DB"), dbConnString(cfg, pass))
	if err != nil {
		return nil, err
	}
	if err := db.DB().Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// DbConnString returns a string to connect to the database
func dbConnString(cfg *config.Config, pass string) (cns string) {
	var sslmode string
	switch cfg.ConnectDB.SslMode {
	case true:
		sslmode = "enable"
	case false:
		sslmode = "disable"
	}
	switch os.Getenv("DIALECT_DB") {
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
