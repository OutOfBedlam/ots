package logging

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/OutOfBedlam/ots/glob"
	"gopkg.in/yaml.v3"
)

type Level int

func (lvl *Level) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}
	*lvl = ParseLogLevel(str)
	return nil
}

func (lvl *Level) UnmarshalJSON(b []byte) error {
	*lvl = ParseLogLevel(string(b))
	return nil
}

func StringToLogLevelHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(LevelInfo) {
		return data, nil
	}
	lvl, flag := ParseLogLevelP(data.(string))
	if flag {
		return lvl, nil
	} else {
		return nil, fmt.Errorf("Invalid log level: %v", data)
	}
}

const (
	LevelAll Level = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

var logLevelNames = []string{"ALL", "TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

func ParseLogLevel(name string) Level {
	n := strings.ToUpper(name)
	switch n {
	default:
		return LevelAll
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "NONE":
		return LevelError + 1
	}
}

func ParseLogLevelP(name string) (Level, bool) {
	n := strings.ToUpper(name)
	switch n {
	default:
		return LevelAll, false
	case "TRACE":
		return LevelTrace, true
	case "DEBUG":
		return LevelDebug, true
	case "INFO":
		return LevelInfo, true
	case "WARN":
		return LevelWarn, true
	case "ERROR":
		return LevelError, true
	case "NONE":
		return LevelError + 1, false
	}
}

func LogLevelName(level Level) string {
	if level >= 0 && int(level) < len(logLevelNames) {
		return logLevelNames[level]
	}
	return "UNKNOWN"
}

type Log interface {
	TraceEnabled() bool
	Trace(string)
	Tracef(fmt string, args ...interface{})
	DebugEnabled() bool
	Debug(string)
	Debugf(fmt string, args ...interface{})
	InfoEnabled() bool
	Info(string)
	Infof(fmt string, args ...interface{})
	WarnEnabled() bool
	Warn(string)
	Warnf(fmt string, args ...interface{})
	ErrorEnabled() bool
	Error(string)
	Errorf(fmt string, args ...interface{})

	Log(level Level, m string)
	Logf(level Level, fmt string, args ...interface{})

	SetLevel(level Level)
	Level() Level
	// PrefixWidth() int
	// SetPrefixWidth(width int)
}

type levelLog struct {
	name        string
	level       Level
	underlying  *log.Logger
	prefixWidth int
}

func (l *levelLog) SetLevel(level Level) { l.level = level }
func (l *levelLog) Level() Level         { return l.level }

func (l *levelLog) TraceEnabled() bool { return l.level >= LevelTrace }
func (l *levelLog) DebugEnabled() bool { return l.level >= LevelDebug }
func (l *levelLog) InfoEnabled() bool  { return l.level >= LevelInfo }
func (l *levelLog) WarnEnabled() bool  { return l.level >= LevelWarn }
func (l *levelLog) ErrorEnabled() bool { return l.level >= LevelError }

func (l *levelLog) Trace(m string) { l.Log(LevelTrace, m) }
func (l *levelLog) Debug(m string) { l.Log(LevelDebug, m) }
func (l *levelLog) Info(m string)  { l.Log(LevelInfo, m) }
func (l *levelLog) Warn(m string)  { l.Log(LevelWarn, m) }
func (l *levelLog) Error(m string) { l.Log(LevelError, m) }

func (l *levelLog) Tracef(fmt string, args ...interface{}) { l.Logf(LevelTrace, fmt, args...) }
func (l *levelLog) Debugf(fmt string, args ...interface{}) { l.Logf(LevelDebug, fmt, args...) }
func (l *levelLog) Infof(fmt string, args ...interface{})  { l.Logf(LevelInfo, fmt, args...) }
func (l *levelLog) Warnf(fmt string, args ...interface{})  { l.Logf(LevelWarn, fmt, args...) }
func (l *levelLog) Errorf(fmt string, args ...interface{}) { l.Logf(LevelError, fmt, args...) }

func (l *levelLog) PrefixWidth() int { return l.prefixWidth }
func (l *levelLog) SetPrefixWidth(width int) {
	if width > 0 {
		l.prefixWidth = width
	} else {
		l.prefixWidth = prefixWidthDefault
	}
}

func (l *levelLog) Log(lvl Level, m string) {
	l.Logf(lvl, "%s", m)
}

func (l *levelLog) Logf(lvl Level, f string, args ...interface{}) {
	if lvl < l.level {
		return
	}
	fnew := fmt.Sprintf("%-5s %s", logLevelNames[lvl], f)
	l.underlying.Printf(fnew, args...)
}

/////////////////////////////////////////////
//
var levelConfig = make(map[string]Level)
var levelDefault = LevelInfo
var prefixWidthDefault = 18

func SetDefaultLevel(lvl Level) {
	levelDefault = lvl
}

func DefaultLevel() Level {
	return levelDefault
}

func SetDefaultPrefixWidth(width int) {
	if width > 0 {
		prefixWidthDefault = width
	} else {
		prefixWidthDefault = 18
	}
}

func DefaultPrefixWidth() int {
	return prefixWidthDefault
}

func SetLevel(name string, lvl Level) {
	levelConfig[name] = lvl
}

func GetLevel(name string) Level {
	var matchedPattern string
	var matchedLevel Level

	for pattern, level := range levelConfig {
		if match, err := glob.Match(pattern, name); match && err == nil {
			if len(matchedPattern) < len(pattern) {
				matchedPattern = pattern
				matchedLevel = level
			}
		}
	}

	if matchedPattern != "" {
		return matchedLevel
	}

	return levelDefault
}
