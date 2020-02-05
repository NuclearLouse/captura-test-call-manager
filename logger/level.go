package logger

type Level int

//level constants
// higher level means more serious : DEBUG < INFO < WARN < ERROR < FATAL
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "**TRACE**  "
	case LevelDebug:
		return "**DEBUG**  "
	case LevelInfo:
		return "**STATUS** "
	case LevelWarn:
		return "**WARNING**"
	case LevelError:
		return "**ERROR**  "
	case LevelPanic:
		return "**FAILURE**"
	case LevelFatal:
		return "**FATAL**  "
	default:
		return "<NA>"
	}
}
