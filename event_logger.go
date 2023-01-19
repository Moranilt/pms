package pms

import (
	"os"
	"strings"
)

var (
	yellowBg = "\033[33m"
	redBg    = "\033[31m"
	green    = "\033[32m"
	reset    = "\033[0m"
)

type LogType string

const (
	LOG_INFO  LogType = "info"
	LOG_ERROR LogType = "error"
	LOG_WARN  LogType = "warn"
)

type Logger interface {
	Info(message ...string)
	Error(message ...string)
	Warn(message ...string)
}

type eventLogger struct {
}

func newEventLogger() Logger {
	return &eventLogger{}
}

func (l *eventLogger) formatString(lt LogType, message ...string) string {
	var s strings.Builder
	switch lt {
	case LOG_INFO:
		s.WriteString(green)
	case LOG_ERROR:
		s.WriteString(redBg)
	case LOG_WARN:
		s.WriteString(yellowBg)
	}
	s.WriteString(strings.Join(message, " "))
	s.WriteString(reset)
	s.WriteString("\n")
	return s.String()
}

func (l *eventLogger) Info(message ...string) {
	os.Stdout.WriteString(l.formatString(LOG_INFO, message...))
}

func (l *eventLogger) Error(message ...string) {
	os.Stdout.WriteString(l.formatString(LOG_ERROR, message...))
}

func (l *eventLogger) Warn(message ...string) {
	os.Stdout.WriteString(l.formatString(LOG_WARN, message...))
}
