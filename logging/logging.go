package logging

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

//////////////////////////////////////////////////////////////////////

type Logging interface {
	Name() string
	Printf(format string, v ...interface{})
	Write(p []byte) (n int, err error)
}

type Config struct {
	Name               string        `yaml:"name" json:"name"`
	Console            bool          `yaml:"console" json:"console"`
	Filename           string        `yaml:"filename" json:"filename"`
	Append             bool          `yaml:"append" json:"append"`
	RotateSchedule     string        `yaml:"roateSchedule" json:"rotateSchedule"`
	MaxSize            int           `yaml:"maxSize" json:"maxSize"`
	MaxBackups         int           `yaml:"maxBackups" json:"maxBackups"`
	MaxAge             int           `yaml:"maxAge" json:"maxAge"`
	Compress           bool          `yaml:"compress" json:"compress"`
	Levels             []LevelConfig `yaml:"levels" json:"levels"`
	DefaultPrefixWidth int           `yaml:"defaultPrefixWidth" json:"defaultPrefixWidth"`
	DefaultLevel       Level         `yaml:"defaultLevel" json:"defaultLevel"`
}

type LevelConfig struct {
	Pattern string `yaml:"pattern" json:"pattern"`
	Level   Level  `yaml:"level" json:"level"`
}

//////////////////////////////////////////////////////////////////////

// Logger
var defaultLogging Logging

// alternative defaultWriter when Logger is not configured
var defaultWriter io.Writer

// Discard logger used when no `Logger` and `writer` configures
var Discard = log.New(ioutil.Discard, "", 0)

var rotateCron = cron.New()

func Configure(cfg *Config) Logging {
	defaultLogging = New(cfg)
	for _, c := range cfg.Levels {
		levelConfig[c.Pattern] = c.Level
	}
	SetDefaultPrefixWidth(cfg.DefaultPrefixWidth)
	SetDefaultLevel(cfg.DefaultLevel)
	return defaultLogging
}

func ConfigureWithContent(str []byte) error {
	cfg := Config{}
	err := yaml.Unmarshal(str, &cfg)
	if err != nil {
		return err
	}

	Configure(&cfg)
	return nil
}

func ConfigureWithString(str string) error {
	return ConfigureWithContent([]byte(str))
}

func ConfigureWithFile(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ConfigureWithContent(content)
}

func New(cfg *Config) Logging {
	if cfg.Filename == "." {
		return NewLoggingWithDiscard(cfg.Name)
	} else if cfg.Filename == "-" {
		return NewLoggingWithWriter(cfg.Name, os.Stdout)
	} else {
		lj := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}
		logger := NewLoggingWithWriter(cfg.Name, lj)
		logger.(*LogAppender).console = cfg.Console
		if !cfg.Append {
			lj.Rotate()
		}
		if len(cfg.RotateSchedule) > 0 {
			rotateCron.AddFunc(cfg.RotateSchedule, func() {
				lj.Rotate()
			})
			go rotateCron.Run()
		}
		return logger
	}
}

func SetDefaultLogging(l Logging) Logging {
	old := defaultLogging
	defaultLogging = l
	return old
}

func SetDefaultWriter(out io.Writer) {
	defaultWriter = out
}

func NewLoggingWithDiscard(name string) Logging {
	return NewLoggingWithWriter(name, ioutil.Discard)
}

func NewLoggingWithWriter(name string, out io.Writer) Logging {
	return &LogAppender{
		name: name,
		out:  out,
	}
}

func NewWithFilename(name string, filename string, maxSize int, console bool) Logging {
	lj := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
		LocalTime:  true,
	}
	logger := NewLoggingWithWriter(name, lj)
	logger.(*LogAppender).console = console
	return logger
}

func GetLogger(prefix string) *log.Logger {
	return GetLoggerOpt(prefix, log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
}

func GetLoggerOpt(prefix string, opt int) *log.Logger {
	f := fmt.Sprintf("%%-%ds ", DefaultPrefixWidth())
	fprefix := fmt.Sprintf(f, prefix)

	if defaultLogging == nil {
		if defaultWriter != nil {
			return log.New(defaultWriter, fprefix, 0)
		} else {
			return Discard
		}
	}
	return log.New(defaultLogging, fprefix, opt)
}

func GetLog(name string) Log {
	return GetLogOpt(name, log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
}

func GetLogOpt(name string, opt int) Log {
	level := GetLevel(name)
	pfmt := fmt.Sprintf("%%-%ds ", prefixWidthDefault)
	//prefix := fmt.Sprintf(pfmt, "["+name+"]")
	prefix := fmt.Sprintf(pfmt, name)
	var underlying *log.Logger
	if defaultLogging == nil {
		if defaultWriter != nil {
			underlying = log.New(defaultWriter, prefix, 0)
		} else {
			underlying = Discard
		}
	} else {
		underlying = log.New(defaultLogging, prefix, opt)
	}

	return &levelLog{
		name:        name,
		level:       level,
		underlying:  underlying,
		prefixWidth: prefixWidthDefault,
	}
}

func LogError(prefix string, action string, err error) {
	defaultLogging.Printf("[%s] error during %s (%s)", prefix, action, err.Error())
}

func LogErrorf(prefix string, err error, format string, args ...interface{}) {
	defaultLogging.Printf("[%s] error during %s (%s)", prefix, fmt.Sprintf(format, args...), err.Error())
}

func LogAction(prefix string, action string) {
	defaultLogging.Printf("[%s] %s", prefix, action)
}

func LogActionf(prefix string, format string, args ...interface{}) {
	defaultLogging.Printf("[%s] %s", prefix, fmt.Sprintf(format, args...))
}

func LogTarget(prefix string, action string, target interface{}) {
	defaultLogging.Printf("[%s] %s (%v)", prefix, action, target)
}
