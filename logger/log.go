package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	defaultTimeFormat   = "15:04:05"
	defaultPrefixFormat = "%s|%s|%s|%s" //time, code ,lvl, message
	// defaultPrefixFormat = "%s : %.3s : %s"
)

var writer *RotateLogger

type errCode int

// func (ec errCode) String() string {
// 	c := fmt.Sprintf("%d", ec)
// 	return strings.Repeat("0", 5-len(c)) + c
// }

func (ec errCode) String() string {
	return fmt.Sprintf("%05d", ec)
}

func Trace(v ...interface{}) {
	output(LevelTrace, "00000", fmt.Sprintln(v...))
}

func Tracef(format string, v ...interface{}) {
	output(LevelTrace, "00000", fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Debug(v ...interface{}) {
	output(LevelDebug, "00000", fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
	output(LevelDebug, "00000", fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Info(v ...interface{}) {
	output(LevelInfo, "00000", fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	output(LevelInfo, "00000", fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Warn(v ...interface{}) {
	output(LevelWarn, "00000", fmt.Sprintln(v...))
}

func Warnf(format string, v ...interface{}) {
	output(LevelWarn, "00000", fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Error(ec errCode, v ...interface{}) {
	output(LevelError, ec.String(), fmt.Sprintln(v...))
}

func Errorf(ec errCode, format string, v ...interface{}) {
	output(LevelError, ec.String(), fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Panic(ec errCode, v ...interface{}) {
	output(LevelPanic, ec.String(), fmt.Sprintln(v...))
}

func Panicf(ec errCode, format string, v ...interface{}) {
	output(LevelPanic, ec.String(), fmt.Sprintln(fmt.Sprintf(format, v...)))
}

func Fatal(ec errCode, v ...interface{}) {
	output(LevelFatal, ec.String(), fmt.Sprintln(v...))
	writer.Close()
	os.Exit(1)
}

func Fatalf(ec errCode, format string, v ...interface{}) {
	output(LevelFatal, ec.String(), fmt.Sprintln(fmt.Sprintf(format, v...)))
	writer.Close()
	os.Exit(1)
}

func getLevelRotate(lvl, rotateRule string) (Level, int) {
	var loglevel Level
	var rotate int
	switch strings.ToLower(lvl) {
	case "trace", "trac":
		loglevel = LevelTrace
	case "debug", "deb":
		loglevel = LevelDebug
	case "info", "inf":
		loglevel = LevelInfo
	case "warning", "warn":
		loglevel = LevelWarn
	case "error", "err":
		loglevel = LevelError
	case "panic", "pan":
		loglevel = LevelPanic
	case "fatal", "fat":
		loglevel = LevelFatal
	default:
		loglevel = -1
	}

	switch strings.ToLower(rotateRule) {
	case "monthly", "month":
		rotate = MonthlyRotate
	case "daily", "day":
		rotate = DailyRotate
	case "hourly", "hour":
		rotate = HourlyRotate
	case "minutely", "minut":
		rotate = MinutelyRotate
	default:
		rotate = -1
	}
	return loglevel, rotate
}

func Setup(path, lvl, rotateRule string) error {

	level, rotateType := getLevelRotate(lvl, rotateRule)

	if level < LevelTrace || level > LevelFatal {
		return errors.New("None Exist Level")
	}

	if rotateType < MonthlyRotate || rotateType > MinutelyRotate {
		return errors.New("None Exist Rotate Type")
	}
	var err error
	if writer, err = NewRotateLogger(path, level, rotateType); err != nil {
		return err
	}
	return nil
}

func SetLevel(level Level) {
	writer.SetLevel(level)
}

func GetLevel() Level {
	return writer.level
}

func output(level Level, code, content string) {
	if level < writer.level {
		return
	}

	//the writer may be close
	// defaultPrefixFormat = "%s|%s|%s|%s" //time, code ,lvl, message
	logContent := fmt.Sprintf(defaultPrefixFormat, time.Now().Format(defaultTimeFormat), code, level.String(), content)
	if writer != nil {
		buf := make([]byte, len(logContent))
		copy(buf, logContent)
		writer.Write(buf)
	} else {
		log.Print(logContent)
	}
}

func Close() error {
	if writer != nil {
		if err := writer.Close(); err != nil {
			return err
		} else {
			writer = nil
		}
	}
	return nil
}
