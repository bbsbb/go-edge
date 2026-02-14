//go:build testing

package testing

import (
	"io"
	"log/slog"
	"os"

	"github.com/stretchr/testify/suite"
)

type LogCapture struct {
	writer *os.File
	reader *os.File
	Logger *slog.Logger
	suite  suite.TestingSuite
}

func (lc *LogCapture) Stop() {
	if err := lc.writer.Close(); err != nil {
		lc.suite.T().Fatalf("Log capture writer pipe could not be closed: %v", err)
	}
}

func (lc *LogCapture) Output() string {
	lc.Stop()

	captured, err := io.ReadAll(lc.reader)
	if err != nil {
		lc.suite.T().Fatalf("Could not read captured logs: %v", err)
	}
	return string(captured)
}

// NewNoopLogger returns a logger that discards all output.
func NewNoopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func NewLogCapture(s suite.TestingSuite) *LogCapture {
	r, w, err := os.Pipe()
	if err != nil {
		s.T().Fatalf("Could not instantiate log capture: %v", err)
	}

	return &LogCapture{
		suite:  s,
		Logger: slog.New(slog.NewTextHandler(w, nil)),
		reader: r,
		writer: w,
	}
}
