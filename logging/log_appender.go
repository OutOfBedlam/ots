package logging

import (
	"fmt"
	"io"
	"os"
)

type LogAppender struct {
	name    string
	out     io.Writer
	console bool
}

//////////////////////////////////////////
// BEGIN: implements logging.Logging
func (s *LogAppender) Name() string {
	return s.name
}

func (s *LogAppender) Printf(format string, v ...interface{}) {
	if s.console {
		fmt.Printf(format, v...)
	}
	s.Write([]byte(fmt.Sprintf(format+"\n", v...)))
}

func (s *LogAppender) Write(p []byte) (n int, err error) {
	if s.console {
		os.Stdout.Write(p)
	}
	return s.out.Write(p)
}

// END: implements logging.Logging
//////////////////////////////////////////

func (s *LogAppender) Configure(cfg *Config) error {
	return nil
}
