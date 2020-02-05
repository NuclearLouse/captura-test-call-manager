package logger

import (
	"time"
)

const (
	MonthlyRotate = iota
	DailyRotate
	HourlyRotate
	MinutelyRotate
)

const (
	monthlyDateFormat    = "200601"
	dailyDateFormat      = "20060102"
	hourlyDateFormat     = "20060102T15"
	minutelyRotateFormat = "20060102T15-04"
)

type RotateRule struct {
	rotateType int
	rotateTime string
}

func NewRotateRule(rotateType int) *RotateRule {
	return &RotateRule{
		rotateType: rotateType,
		rotateTime: getFormatDate(rotateType),
	}
}

func (rr *RotateRule) ShallRotate() bool {
	return rr.rotateTime != getFormatDate(rr.rotateType) && len(rr.rotateTime) > 0
}

func (rr *RotateRule) SetRotateTime() {
	rr.rotateTime = getFormatDate(rr.rotateType)
}

// func (rr *RotateRule) GetBackupFilename(filename string) string {
// 	f := strings.Split(filename, "_")
// 	return fmt.Sprintf("%s_%s.log", f[0], getFormatDate(rr.rotateType))
// }

func getFormatDate(rotateType int) string {
	switch rotateType {
	case MonthlyRotate:
		return time.Now().Format(monthlyDateFormat)
	case DailyRotate:
		return time.Now().Format(dailyDateFormat)
	case HourlyRotate:
		return time.Now().Format(hourlyDateFormat)
	case MinutelyRotate:
		return time.Now().Format(minutelyRotateFormat)
	default:
		return time.Now().Format(dailyDateFormat)
	}
}
