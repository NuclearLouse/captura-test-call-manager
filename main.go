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
	"runtime"
	"syscall"

	"redits.oculeus.com/asorokin/my_packages/crypter"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

const (
	cfgFile = "CaptTestCallsSrvc.ini"
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
	switch runtime.GOOS {
	case "windows":
		os.Setenv("DURATION", cfg.Decoders.Ffmpeg+" -i %s 2>&1 | findstr Duration")
		os.Setenv("WAV_FORM_IMG", cfg.Decoders.Ffmpeg+" -i %s -filter_complex [0:a]aformat=channel_layouts=mono,compand=gain=-6,showwavespic=s=500x100:colors=#000000[fg];color=s=500x100:color=#FFFFFF,drawgrid=width=iw/10:height=ih/5:color=#000000@0.1[bg];[bg][fg]overlay=format=rgb,drawbox=x=(iw-w)/2:y=(ih-h)/2:w=iw:h=1:color=#000000 -vframes 1 %s")
	default:
		os.Setenv("DURATION", cfg.Decoders.Ffmpeg+" -i %s 2>&1 | grep Duration")
		os.Setenv("WAV_FORM_IMG", cfg.Decoders.Ffmpeg+` -i %s -filter_complex "[0:a]aformat=channel_layouts=mono,compand=gain=-6,showwavespic=s=500x100:colors=#000000[fg];color=s=500x100:color=#FFFFFF,drawgrid=width=iw/10:height=ih/5:color=#000000@0.1[bg];[bg][fg]overlay=format=rgb,drawbox=x=(iw-w)/2:y=(ih-h)/2:w=iw:h=1:color=#000000" -vframes 1 %s`)
	}
	os.Setenv("CONCAT_MP3", cfg.Decoders.Ffmpeg+" -i %s -i %s -filter_complex [0:a][1:a]concat=n=2:v=0:a=1 %s")
	os.Setenv("DECODE_TO_WAV", cfg.Decoders.Ffmpeg+" -y -i %s %s")
	os.Setenv("FORMAT_IMG", cfg.Application.FormatIMG) // не надо
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
