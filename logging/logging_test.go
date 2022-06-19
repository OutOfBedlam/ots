package logging_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	. "github.com/OutOfBedlam/ots/logging"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var sampleYaml = []byte(`
name: "test-log"
console: true
filename: "test-filename"
append: true
maxSize: 1
maxBackups: 2
maxAge: 3
compress: true
defaultPrefixWidth: 12
defaultLevel: INFO
levels:
  - pattern: "debug_*"
    level: DEBUG
  - pattern: "info_*"
    level: INFO
`)

func assertSampleYaml(t *testing.T, cfg *Config) {
	assert.Equal(t, "test-log", cfg.Name)
	assert.Equal(t, true, cfg.Console)
	assert.Equal(t, "test-filename", cfg.Filename)
	assert.Equal(t, true, cfg.Append)
	assert.Equal(t, 1, cfg.MaxSize)
	assert.Equal(t, 2, cfg.MaxBackups)
	assert.Equal(t, 3, cfg.MaxAge)
	assert.Equal(t, true, cfg.Compress)
	assert.Equal(t, 12, cfg.DefaultPrefixWidth)
	assert.Equal(t, LevelInfo, cfg.DefaultLevel)
	assert.Equal(t, 2, len(cfg.Levels))
	assert.Equal(t, "debug_*", cfg.Levels[0].Pattern)
	assert.Equal(t, LevelDebug, cfg.Levels[0].Level)
	assert.Equal(t, "info_*", cfg.Levels[1].Pattern)
	assert.Equal(t, LevelInfo, cfg.Levels[1].Level)
}

func TestLogConfigContent(t *testing.T) {
	err := ConfigureWithContent(sampleYaml)
	assert.Nil(t, err)
}

func TestLogConfig(t *testing.T) {
	cfg := Config{}
	err := yaml.Unmarshal(sampleYaml, &cfg)
	assert.Nil(t, err)
	assertSampleYaml(t, &cfg)
}

func TestLogConfigUnmarshal(t *testing.T) {
	var cfg = Config{}

	var vi = viper.New()
	vi.SetConfigType("yaml")
	vi.ReadConfig(bytes.NewBuffer(sampleYaml))
	decoderHook := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			StringToLogLevelHookFunc,
			// Function to support net.IP
			//mapstructure.StringToIPHookFunc(),
			// Appended by the two default functions
			//mapstructure.StringToTimeDurationHookFunc(),
			//mapstructure.StringToSliceHookFunc(","),
		))
	err := vi.Unmarshal(&cfg, decoderHook)

	assert.Nil(t, err)
	assertSampleYaml(t, &cfg)
}

func TestLogAction(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger := NewLoggingWithWriter("buffer", buffer)
	org := SetDefaultLogging(logger)
	defer func(l Logging) { SetDefaultLogging(l) }(org)

	LogAction("a", "b")
	assert.Equal(t, "[a] b\n", buffer.String())
}

func TestLogError(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger := NewLoggingWithWriter("buffer", buffer)
	org := SetDefaultLogging(logger)
	defer func(l Logging) { SetDefaultLogging(l) }(org)

	LogError("a", "b", errors.New("err"))
	assert.Equal(t, "[a] error during b (err)\n", buffer.String())
}

func TestLogTarget(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger := NewLoggingWithWriter("buffer", buffer)
	org := SetDefaultLogging(logger)
	defer func(l Logging) { SetDefaultLogging(l) }(org)

	LogTarget("a", "b", 123)
	assert.Equal(t, "[a] b (123)\n", buffer.String())
}

func TestStdErrLogger(t *testing.T) {
	l := NewLoggingWithWriter("stderr", os.Stderr)
	assert.Equal(t, "stderr", l.Name())
}

func TestLogLevel(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger := NewLoggingWithWriter("buffer", buffer)
	org := SetDefaultLogging(logger)
	defer func(l Logging) { SetDefaultLogging(l) }(org)

	SetDefaultPrefixWidth(12)
	SetDefaultLevel(LevelDebug)

	l := GetLog("test.logger")

	l.Debug("log test")
	//2021/11/02 12:06:49.904946 test.logger  DEBUG log test\n
	timelen := len("2021/11/02 12:06:49.904946 ")
	assert.Equal(t, "test.logger  DEBUG log test\n", buffer.String()[timelen:])
}
