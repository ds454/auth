package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	out *os.File
}

var (
	instance *Logger
	once     sync.Once
)

func Get() *Logger {
	once.Do(func() {
		instance = &Logger{out: os.Stdout}
	})
	return instance
}

func (l *Logger) logf(level, format string, args ...any) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file, line = "???", 0
	}

	ts := time.Now().Format("15:04:05")
	color := levelColor(level)
	fmt.Fprintf(l.out, "%s[%s]%s\t%s[%s] %s:%d%s %s\n",
		color, level, colorReset,
		colorGray, ts, filepath.Base(file), line, colorReset,
		fmt.Sprintf(format, args...),
	)
}

func (l *Logger) info(format string, args ...any)  { l.logf("INFO", format, args...) }
func (l *Logger) warn(format string, args ...any)  { l.logf("WARN", format, args...) }
func (l *Logger) error(format string, args ...any) { l.logf("ERROR", format, args...) }
func (l *Logger) debug(format string, args ...any) { l.logf("DEBUG", format, args...) }

func Debug(format string, args ...any) { Get().debug(format, args...) }
func Info(format string, args ...any)  { Get().info(format, args...) }
func Warn(format string, args ...any)  { Get().warn(format, args...) }
func Error(format string, args ...any) { Get().error(format, args...) }
func Abort(format string, args ...any) { Get().error(format, args...); os.Exit(1) }
