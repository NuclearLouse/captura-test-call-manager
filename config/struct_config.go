package config

// Config main sections of the config
type Config struct {
	Logger    `ini:"logger"`
	ConnectDB `ini:"connectdb"`
	Decoders  `ini:"decoders"`
	Timetable `ini:"timetable"`
	// ErrorMail `ini:"errormail"`
}

// Logger section logger
type Logger struct {
	LogPath   string `ini:"log_path"`
	LogLevel  string `ini:"log_level"`
	ShortLvl  bool   `ini:"short_lvl"`
	MaxSize   int    `ini:"max_size"`
	MaxAge    int    `ini:"max_age"`
	MaxBackup int    `ini:"max_backup"`
	Compress  bool   `ini:"compress"`
	Localtime bool   `ini:"localtime"`
	RepeatErr int64  `ini:"repeat_err"`
	Rotate    string `ini:"rotate"`
}

// ConnectDB section DB
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

// Decoders section external decoders
type Decoders struct {
	Ffmpeg string `ini:"ffmpeg"`
	Sox    string `ini:"sox"`
}

// Timetable section for settings intervals
type Timetable struct {
	PrepareRequest       bool  `ini:"prepare_request"`
	IntervalDeleteTests  int64 `ini:"delete_old_tests"`
	IntervalCheckTests   int64 `ini:"check_tests"`
	IntervalPrepareTests int64 `ini:"prepare_tests"`
}

// ErrorMail section for settings error mail
type ErrorMail struct {
	SendMail     bool   `ini:"sendmail"`
	Host         string `ini:"host"`
	Port         int    `ini:"port"`
	EmailFrom    string `ini:"email_from"`
	User         string `ini:"user"`
	Pass         string `ini:"hash"`
	EmailTo      string `ini:"email_to"`
	Headline     string `ini:"headline"`
	SendInterval int64  `ini:"send_interval"`
}
