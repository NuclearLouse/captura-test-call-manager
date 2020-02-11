// Test Calls Management Service
// Home page: https://redits.oculeus.com/asorokin/captura_tcm
// This service integrates NetSense, iTest, Assure test calls systems into Captura System
// NetSense: https://arptel.com/index.html
// iTest: http://www.i-test.net/
// Assure: https://www.csgi.com/portfolio/csg-wholesale/assure/
//
//
// The file contains the function necessary when you first start the program.
// Which reads the config, creates service directories, connects to the database
// and creates, if necessary, tables, sets global variables and starts the service itself.
//
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"./crypter"
	log "./logger"
)

const (
	cfgFile = "CaptTestCallsSrvc.ini"
	logFile = "CaptTestCallsSrvc"
	tmpDir  = "temp"
	key     = "XContextToStoreX"
)

var (
	srvTmpFolder     string
	schemaPG         string
	dialectDB        string
	ffmpegWavFormImg string
	ffmpegDuration   string
	ffmpegConcatMP3  string
	ffmpegDecode     string
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

	logPath := cfg.Logger.LogPath
	if !strings.HasSuffix(logPath, "/") {
		logPath = logPath + "/"
	}
	if err := log.Setup(logPath+logFile, cfg.Logger.LogLevel, cfg.Logger.Rotate); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer func() {
		log.Close()
	}()

	// setting global variables
	srvTmpFolder = appPath + tmpDir + slash
	setGlobalVars(cfg)

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

	go runService(cfg, db)

	waitForSignal()
}

// SetGlobalVars sets the environment variables necessary for work
func setGlobalVars(cfg *Config) {
	switch runtime.GOOS {
	case "windows":
		ffmpegDuration = cfg.Application.Ffmpeg + " -i %s 2>&1 | findstr Duration"
		ffmpegWavFormImg = cfg.Application.Ffmpeg + " -i %s -filter_complex [0:a]aformat=channel_layouts=mono,compand=gain=-6,showwavespic=s=500x100:colors=#000000[fg];color=s=500x100:color=#FFFFFF,drawgrid=width=iw/10:height=ih/5:color=#000000@0.1[bg];[bg][fg]overlay=format=rgb,drawbox=x=(iw-w)/2:y=(ih-h)/2:w=iw:h=1:color=#000000 -vframes 1 %s"
	default:
		ffmpegDuration = cfg.Application.Ffmpeg + " -i %s 2>&1 | grep Duration"
		ffmpegWavFormImg = cfg.Application.Ffmpeg + ` -i %s -filter_complex "[0:a]aformat=channel_layouts=mono,compand=gain=-6,showwavespic=s=500x100:colors=#000000[fg];color=s=500x100:color=#FFFFFF,drawgrid=width=iw/10:height=ih/5:color=#000000@0.1[bg];[bg][fg]overlay=format=rgb,drawbox=x=(iw-w)/2:y=(ih-h)/2:w=iw:h=1:color=#000000" -vframes 1 %s`
	}

	ffmpegConcatMP3 = cfg.Application.Ffmpeg + " -i %s -i %s -filter_complex [0:a][1:a]concat=n=2:v=0:a=1 %s"
	ffmpegDecode = cfg.Application.Ffmpeg + " -y -i %s %s"

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
	schemaPG = schema
	dialectDB = dialect
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
	log.Fatal(11111, "Exit the program. Reason: got signal", s)
}
