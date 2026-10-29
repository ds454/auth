package log

import (
	"fmt"
	"os"
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

func (l *Logger) log(level, msg string) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	color := levelColor(level)
	fmt.Fprintf(l.out, "%s[%s]%s\t%s[%s]%s %s\n",
		color, level, colorReset,
		colorGray, ts, colorReset,
		msg,
	)
}

func (l *Logger) Info(format string, args ...any)  { l.log("INFO", fmt.Sprintf(format, args...)) }
func (l *Logger) Warn(format string, args ...any)  { l.log("WARN", fmt.Sprintf(format, args...)) }
func (l *Logger) Error(format string, args ...any) { l.log("ERROR", fmt.Sprintf(format, args...)) }
func (l *Logger) Debug(format string, args ...any) { l.log("DEBUG", fmt.Sprintf(format, args...)) }

func Debug(format string, args ...any) { Get().Debug(format, args...) }
func Info(format string, args ...any)  { Get().Info(format, args...) }
func Warn(format string, args ...any)  { Get().Warn(format, args...) }
func Error(format string, args ...any) { Get().Error(format, args...) }
func Abort(format string, args ...any) { Get().Error(format, args...); os.Exit(1) }
