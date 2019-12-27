package main

import (
	"errors"
	"fmt"

	"github.com/go-ini/ini"
	"redits.oculeus.com/asorokin/my_packages/crypter"
)

type Config struct {
	Application `ini:"application"`
	Logger      `ini:"logger"`
	ConnectDB   `ini:"connectdb"`
	Decoders    `ini:"decoders"`
	Timetable   `ini:"timetable"`
}

type Application struct {
	FormatIMG      string `ini:"format_img"`
	PrepareRequest bool   `ini:"prepare_request"`
}

type Logger struct {
	LogPath  string `ini:"log_path"`
	LogLevel string `ini:"log_level"`
	Rotate   string `ini:"rotate"`
}

type ConnectDB struct {
	Dialect      string `ini:"db_dialect"`
	Host         string `ini:"host"`
	Port         string `ini:"port"`
	Database     string `ini:"database"`
	SchemaPG     string `ini:"schema_pg"`
	User         string `ini:"user"`
	Pass         string `ini:"password"`
	CryptPass    string `ini:"crypt_pass"`
	SslMode      bool   `ini:"ssl_mode"`
	SQLitePath   string `ini:"sqlite_path"`
	CreateTables bool   `ini:"create_tables"`
}

type Decoders struct {
	Ffmpeg string `ini:"ffmpeg"`
	Sox    string `ini:"sox"`
}

type Timetable struct {
	IntervalDeleteTests  int64 `ini:"delete_old_tests"`
	IntervalCheckTests   int64 `ini:"check_tests"`
	IntervalPrepareTests int64 `ini:"prepare_tests"`
}

func readConfig(configFile, key string) (*Config, error) {
	cfg := &Config{}
	set, err := loadCfgFile(configFile)
	if err != nil {
		return nil, err
	}

	passDB := set.Section("connectdb").Key("password")
	cryptDB := set.Section("connectdb").Key("crypt_pass")
	if passDB.String() == "" && cryptDB.String() == "" {
		err := errors.New("the password for connecting to the database is not set")
		return nil, err
	}

	if passDB.String() != "" {
		ciphertext, err := crypter.Encrypt([]byte(key), []byte(passDB.String()))
		if err == nil {
			cryptDB.SetValue(fmt.Sprintf("%x", ciphertext))
			passDB.SetValue("")
		}

		set.SaveTo(configFile)

		err = set.Reload()
		if err != nil {
			return nil, err
		}
	}

	if err = set.MapTo(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadCfgFile(configFile string) (*ini.File, error) {
	f, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configFile)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func rewriteConfig(configFile, section, key, value string) error {
	cfg, err := loadCfgFile(configFile)
	if err != nil {
		return err
	}
	cfg.Section(section).Key(key).SetValue(value)
	if err := cfg.SaveTo(configFile); err != nil {
		return err
	}
	return nil
}
