package migrations

import (
	"fmt"
	"log/slog"
)

type gooseLogger struct {
	logger *slog.Logger
}

func newGooseLogger(logger *slog.Logger) *gooseLogger {
	return &gooseLogger{logger: logger}
}

func (l gooseLogger) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

func (l gooseLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...))
}
