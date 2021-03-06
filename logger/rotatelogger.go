package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	maxWaitTime          = 100 * time.Millisecond
	bufferSize           = 100
	defaultFileMode      = 0600
	defaultDirectoryMode = 0755
)

// RotateLogger ...
type RotateLogger struct {
		filename       string
		backupFilename string
		level          Level
		rule           *RotateRule
		fp             *os.File
		msg            chan []byte
		done           chan bool
		waitGroup      sync.WaitGroup
	}


func newFileName(filename string, rotateType int) string {
	absPath := path.Dir(filename) + "/"
	f := strings.Split(path.Base(filename), "_")
	return fmt.Sprintf("%s%s_%s.log", absPath, f[0], getFormatDate(rotateType))
}

// NewRotateLogger ...
func NewRotateLogger(filename string, level Level, rotateType int) (*RotateLogger, error) {
	f := fmt.Sprintf("%s_%s.log", filename, getFormatDate(rotateType))
	l := &RotateLogger{
		filename: f,
		rule:     NewRotateRule(rotateType),
		level:    level,
		msg:      make(chan []byte, bufferSize),
		done:     make(chan bool),
	}

	if err := l.init(); err != nil {
		return nil, err
	}

	l.run()
	return l, nil
}

func (rl *RotateLogger) init() error {
	rl.backupFilename = rl.filename
	// rl.backupFilename = rl.rule.GetBackupFilename(rl.filename)
	// fmt.Println(rl.backupFilename)
	if _, err := os.Stat(rl.filename); err != nil {
		basePath := path.Dir(rl.filename)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, defaultDirectoryMode); err != nil {
				return err
			}
		}
		if rl.fp, err = os.Create(rl.filename); err != nil {
			return err
		}
	} else if rl.fp, err = os.OpenFile(rl.filename, os.O_APPEND|os.O_WRONLY, defaultFileMode); err != nil {
		return err
	}
	return nil
}

func (rl *RotateLogger) run() {
	rl.waitGroup.Add(1)

	go func() {
		defer rl.waitGroup.Done()

		for {
			select {
			case msg, ok := <-rl.msg:
				if ok {
					rl.write(msg)
				} else {
					return
				}
			case <-rl.done:
				if len(rl.msg) == 0 {
					return
				}
			}
		}
	}()
}

func (rl *RotateLogger) rotate() error {
	if rl.fp != nil {
		if err := rl.fp.Close(); err != nil {
			return err
		}
		rl.fp = nil
	}

	_, err := os.Stat(rl.filename)
	if err == nil && len(rl.backupFilename) > 0 {
		err = os.Rename(rl.filename, rl.backupFilename)
		// if err != nil {
		// 	return err
		// }
	}

	rl.filename = newFileName(rl.filename, rl.rule.rotateType)
	rl.backupFilename = rl.filename
	// rl.backupFilename = rl.rule.GetBackupFilename(rl.filename)
	rl.fp, err = os.Create(rl.filename)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (rl *RotateLogger) Write(content []byte) (int, error) {
	select {
	case <-rl.done:
		return 0, errors.New("Error: log file closed")
	default:
		select {
		case rl.msg <- content:
			return len(content), nil
		case <-time.After(maxWaitTime):
			return 0, errors.New("Timeout on writting log")
		}

	}
}

func (rl *RotateLogger) write(content []byte) {
	if rl.rule.ShallRotate() {
		if err := rl.rotate(); err != nil {
			log.Println(err)
		} else {
			rl.rule.SetRotateTime()
		}
	}
	if rl.fp != nil {
		rl.fp.Write(content)
	}
}

// SetLevel ...
func (rl *RotateLogger) SetLevel(level Level) {
	rl.level = level
}

// GetLevel ...
func (rl *RotateLogger) GetLevel() Level {
	return rl.level
}

// Close ...
func (rl *RotateLogger) Close() error {
	close(rl.done)
	close(rl.msg)
	rl.waitGroup.Wait()
	if err := rl.fp.Sync(); err != nil {
		return err
	}
	return rl.fp.Close()
}
