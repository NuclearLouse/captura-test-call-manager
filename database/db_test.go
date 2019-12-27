package database

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"

	"redits.oculeus.com/asorokin/tcm/config"
)

func confDB() *config.Config {
	return &config.Config{
		ConnectDB: config.ConnectDB{
			Dialect:  "postgres",
			Host:     "localhost",
			Port:     "5432",
			Database: "postgres",
			SchemaPG: "public",
			User:     "postgres",
			Pass:     "postgres",
		},
	}
}

func TestConnectOptions(t *testing.T) {
	absPath := "C:\\User\\go\\src"
	cfg := confDB()
	gotDialect, gotOptions := connectOptions(cfg, absPath)
	wantOptions := "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"
	if gotOptions != wantOptions {
		t.Fatalf("got %q, wanted %q", gotOptions, wantOptions)
	}
	wantDialect := "postgres"
	if gotDialect != wantDialect {
		t.Fatalf("got %q, wanted %q", gotDialect, wantDialect)
	}
}

func TestConnect(t *testing.T) {
	absPath := "C:\\User\\go\\src"
	cfg := confDB()
	dialect, options := connectOptions(cfg, absPath)
	db, err := gorm.Open(dialect, options)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Exec("SET application_name = 'my_test_app'").Error; err != nil {
		t.Fatal(err)
	}
	type result struct{ ApplicationName string }
	var name result
	if err := db.Raw("SHOW application_name").Scan(&name).Error; err != nil {
		fmt.Println(name)
		t.Fatal(err)
	}
	if name.ApplicationName != "my_test_app" {
		t.Fatalf(`got %q, wanted "my_test_app"`, name.ApplicationName)
	}

}
