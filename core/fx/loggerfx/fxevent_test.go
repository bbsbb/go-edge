package loggerfx

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxevent"
)

type FxEventSuite struct {
	suite.Suite
}

func (s *FxEventSuite) TestFxEvent_String() {
	tests := []struct {
		event FxEvent
		want  string
	}{
		{FxEventOnStart, "onStart"},
		{FxEventOnStop, "onStop"},
		{FxEventProvided, "provided"},
	}

	for _, tt := range tests {
		s.Run(tt.want, func() {
			s.Assert().Equal(tt.want, tt.event.String())
		})
	}
}

func (s *FxEventSuite) TestFxEvent_WithState() {
	tests := []struct {
		name  string
		event FxEvent
		state FxEventState
		want  string
	}{
		{"onStart done", FxEventOnStart, FxStateDone, "onStart.done"},
		{"onStart error", FxEventOnStart, FxStateError, "onStart.error"},
		{"provided done", FxEventProvided, FxStateDone, "provided.done"},
		{"provided error", FxEventProvided, FxStateError, "provided.error"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Assert().Equal(tt.want, tt.event.WithState(tt.state))
		})
	}
}

func (s *FxEventSuite) TestFxEvent_WithAttrs() {
	args := FxEventOnStart.WithAttrs(FxAttrCallee, "func", FxAttrCaller, "caller")
	s.Assert().Equal([]any{"callee", "func", "caller", "caller"}, args)
}

func (s *FxEventSuite) TestFxAttrKey_String() {
	tests := []struct {
		key  FxAttrKey
		want string
	}{
		{FxAttrCallee, "callee"},
		{FxAttrCaller, "caller"},
		{FxAttrModule, "module"},
	}

	for _, tt := range tests {
		s.Run(tt.want, func() {
			s.Assert().Equal(tt.want, tt.key.String())
		})
	}
}

func (s *FxEventSuite) TestNewFxSlogLogger() {
	logger, _ := newLogger(&Configuration{Level: LogLevelInfo, Format: LogFormatJSON}, os.Stdout)
	fxLogger := NewFxSlogLogger(logger)
	s.Assert().NotNil(fxLogger)
}

func (s *FxEventSuite) TestLogEvent() {
	testErr := errors.New("test error")

	tests := []struct {
		name           string
		event          fxevent.Event
		expectedMsg    string
		expectedLevel  string
		expectedFields map[string]any
	}{
		{
			name:          "OnStartExecuting",
			event:         &fxevent.OnStartExecuting{FunctionName: "myFunc", CallerName: "myCaller"},
			expectedMsg:   "onStart",
			expectedLevel: "INFO",
			expectedFields: map[string]any{
				"callee": "myFunc",
				"caller": "myCaller",
			},
		},
		{
			name:          "OnStartExecuted success",
			event:         &fxevent.OnStartExecuted{FunctionName: "myFunc", CallerName: "myCaller"},
			expectedMsg:   "onStart.done",
			expectedLevel: "INFO",
		},
		{
			name:          "OnStartExecuted error",
			event:         &fxevent.OnStartExecuted{FunctionName: "myFunc", CallerName: "myCaller", Err: testErr},
			expectedMsg:   "onStart.error",
			expectedLevel: "ERROR",
			expectedFields: map[string]any{
				"error": "test error",
			},
		},
		{
			name:          "OnStopExecuting",
			event:         &fxevent.OnStopExecuting{FunctionName: "myFunc", CallerName: "myCaller"},
			expectedMsg:   "onStop",
			expectedLevel: "INFO",
		},
		{
			name:          "OnStopExecuted success",
			event:         &fxevent.OnStopExecuted{FunctionName: "myFunc", CallerName: "myCaller"},
			expectedMsg:   "onStop.done",
			expectedLevel: "INFO",
		},
		{
			name:          "OnStopExecuted error",
			event:         &fxevent.OnStopExecuted{FunctionName: "myFunc", CallerName: "myCaller", Err: testErr},
			expectedMsg:   "onStop.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Supplied success",
			event:         &fxevent.Supplied{TypeName: "*Config", ModuleName: "app"},
			expectedMsg:   "supplied.done",
			expectedLevel: "DEBUG",
			expectedFields: map[string]any{
				"type":   "*Config",
				"module": "app",
			},
		},
		{
			name:          "Supplied error",
			event:         &fxevent.Supplied{TypeName: "*Config", ModuleName: "app", Err: testErr},
			expectedMsg:   "supplied.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Provided",
			event:         &fxevent.Provided{ConstructorName: "NewService", ModuleName: "app", OutputTypeNames: []string{"*Service"}},
			expectedMsg:   "provided",
			expectedLevel: "DEBUG",
			expectedFields: map[string]any{
				"type":        "*Service",
				"module":      "app",
				"constructor": "NewService",
			},
		},
		{
			name:          "Provided error",
			event:         &fxevent.Provided{ModuleName: "app", Err: testErr},
			expectedMsg:   "provided.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Replaced",
			event:         &fxevent.Replaced{ModuleName: "app", OutputTypeNames: []string{"*Service"}},
			expectedMsg:   "replaced",
			expectedLevel: "INFO",
			expectedFields: map[string]any{
				"type":   "*Service",
				"module": "app",
			},
		},
		{
			name:          "Replaced error",
			event:         &fxevent.Replaced{ModuleName: "app", Err: testErr},
			expectedMsg:   "replaced.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Decorated",
			event:         &fxevent.Decorated{DecoratorName: "Decorate", ModuleName: "app", OutputTypeNames: []string{"*Service"}},
			expectedMsg:   "decorated",
			expectedLevel: "INFO",
			expectedFields: map[string]any{
				"type":      "*Service",
				"module":    "app",
				"decorator": "Decorate",
			},
		},
		{
			name:          "Decorated error",
			event:         &fxevent.Decorated{ModuleName: "app", Err: testErr},
			expectedMsg:   "decorated.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Invoking",
			event:         &fxevent.Invoking{FunctionName: "main.Run", ModuleName: "app"},
			expectedMsg:   "invoking",
			expectedLevel: "DEBUG",
			expectedFields: map[string]any{
				"function": "main.Run",
				"module":   "app",
			},
		},
		{
			name:          "Invoked error",
			event:         &fxevent.Invoked{FunctionName: "main.Run", ModuleName: "app", Err: testErr},
			expectedMsg:   "invoked.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Run success",
			event:         &fxevent.Run{Name: "task", ModuleName: "app"},
			expectedMsg:   "run.done",
			expectedLevel: "DEBUG",
			expectedFields: map[string]any{
				"name":   "task",
				"module": "app",
			},
		},
		{
			name:          "Run error",
			event:         &fxevent.Run{Name: "task", ModuleName: "app", Err: testErr},
			expectedMsg:   "run.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Stopping",
			event:         &fxevent.Stopping{Signal: os.Interrupt},
			expectedMsg:   "stopping",
			expectedLevel: "INFO",
			expectedFields: map[string]any{
				"signal": "INTERRUPT",
			},
		},
		{
			name:          "Stopped error",
			event:         &fxevent.Stopped{Err: testErr},
			expectedMsg:   "stopped.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "RollingBack",
			event:         &fxevent.RollingBack{StartErr: testErr},
			expectedMsg:   "rollback.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "RolledBack error",
			event:         &fxevent.RolledBack{Err: testErr},
			expectedMsg:   "rollback.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Started success",
			event:         &fxevent.Started{},
			expectedMsg:   "started.done",
			expectedLevel: "INFO",
		},
		{
			name:          "Started error",
			event:         &fxevent.Started{Err: testErr},
			expectedMsg:   "started.error",
			expectedLevel: "ERROR",
		},
		{
			name:          "LoggerInitialized success",
			event:         &fxevent.LoggerInitialized{},
			expectedMsg:   "loggerInit.done",
			expectedLevel: "DEBUG",
		},
		{
			name:          "LoggerInitialized error",
			event:         &fxevent.LoggerInitialized{Err: testErr},
			expectedMsg:   "loggerInit.error",
			expectedLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var buf bytes.Buffer
			logger, _ := newLogger(&Configuration{Level: LogLevelDebug, Format: LogFormatJSON}, &buf)
			fxLogger := NewFxSlogLogger(logger)

			fxLogger.LogEvent(tt.event)

			var entry map[string]any
			s.Require().NoError(json.Unmarshal(buf.Bytes(), &entry))
			s.Assert().Equal(tt.expectedMsg, entry["msg"])
			s.Assert().Equal(tt.expectedLevel, entry["level"])

			for k, v := range tt.expectedFields {
				s.Assert().Equal(v, entry[k], "field %s", k)
			}
		})
	}
}

func TestFxEventSuite(t *testing.T) {
	suite.Run(t, new(FxEventSuite))
}
