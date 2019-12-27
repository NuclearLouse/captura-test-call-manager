// Package main of TCM is the main package defining the entry point
// and which compiles into the tcm or tcm.exe executable.
//
// Home page: https://redits.oculeus.com/asorokin/tcm
//
//
// The file contains the function necessary when you first start the program.
// Which reads the config, creates service directories, connects to the database
// and creates, if necessary, tables.
// The main function itself contains only the start of initialization
// and the start of the main service.
//
package main

import (
	// l "log"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"redits.oculeus.com/asorokin/my_packages/crypter"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

const (
	cfgFile = "tcm.ini"
	tmpDir  = "temp"
	key     = "XContextToStoreX"
)

func main() {
	path, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	absPath := filepath.Dir(path)
	slash := string(os.PathSeparator)
	appPath := absPath + slash

	cfg, err := readConfig(appPath+cfgFile, key)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create dir for service temp files
	if err := createDir(appPath + tmpDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := log.Setup(cfg.Logger.LogPath, cfg.Logger.LogLevel, cfg.Logger.Rotate); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer func() {
		log.Close()
	}()

	if err := setEnvVars(cfg, appPath, slash); err != nil {
		log.Fatalf(0, "Сould not set environment variables|%v", err)
	}
	pass, err := crypter.Decrypt([]byte(key), cfg.ConnectDB.CryptPass)
	if err != nil {
		log.Fatalf(0, "Сould not decrypt password from database|%v", err)
	}
	db, err := newDB(cfg, pass)
	if err != nil {
		log.Fatalf(0, "Сould not connect to database|%v", err)
	}
	if cfg.ConnectDB.CreateTables {
		if err := createTables(db); err != nil {
			os.Exit(1)
		}

		if err := rewriteConfig(appPath+cfgFile, "connectdb", "create_tables", "false"); err != nil {
			log.Errorf(0, "Could not rewrite config file|%v", err)
		}
		fmt.Println("Successfully created tables and overwritten config")
		os.Exit(0)
	}

	go checkOldTests(cfg, db)

	go runService(cfg, db)

	waitForSignal()
}

// SetEnvVars sets the environment variables necessary for work
func setEnvVars(cfg *Config, appPath, slash string) error {
	os.Setenv("SOX", cfg.Decoders.Sox)
	os.Setenv("FFMPEG", cfg.Decoders.Ffmpeg)
	os.Setenv("FORMAT_IMG", cfg.Application.FormatIMG)
	os.Setenv("ABS_PATH_DWL", appPath+tmpDir+slash)

	schema := cfg.ConnectDB.SchemaPG + "."
	if schema == "" {
		schema = "public."
	}
	dialect := cfg.ConnectDB.Dialect
	if dialect == "sqlite" {
		dialect = "sqlite3"
	}
	if dialect == "sqlite3" || dialect == "mysql" {
		schema = ""
	}
	os.Setenv("SCHEMA_PG", schema)
	os.Setenv("DIALECT_DB", dialect)
	return nil
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	s := <-sigChan
	log.Warnf("Exit the program. Got signal %s", s)
}
