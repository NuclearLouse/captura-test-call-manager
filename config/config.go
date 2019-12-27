package config

import (
	"errors"
	"fmt"

	"github.com/go-ini/ini"
	"redits.oculeus.com/asorokin/my_packages/crypter"
)

func loadCfgFile(configFile string) (*ini.File, error) {
	f, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configFile)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ReadConfig read the config file
func ReadConfig(configFile, key string) (*Config, error) {
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

//RewriteConfig overwrites the parameter in the config file
func RewriteConfig(configFile, section, key, value string) error {
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
